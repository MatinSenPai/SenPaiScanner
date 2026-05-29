// Package store persists scan history in a small SQLite leaderboard.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// DB wraps the SQLite connection used for scan history and resume checkpoints.
type DB struct {
	db *sql.DB
}

// Record is a persisted leaderboard row.
type Record struct {
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	ProbeMode     string    `json:"probe_mode"`
	AvgMs         float64   `json:"avg_ms"`
	MinMs         float64   `json:"min_ms"`
	MaxMs         float64   `json:"max_ms"`
	JitterMs      float64   `json:"jitter_ms"`
	LossPct       float64   `json:"loss_pct"`
	ThroughputBPS float64   `json:"throughput_bps"`
	SpeedTested   bool      `json:"speed_tested"`
	Colo          string    `json:"colo"`
	TLSOk         bool      `json:"tls_ok"`
	HTTPStatus    int       `json:"http_status"`
	Score         float64   `json:"score"`
	BestScore     float64   `json:"best_score"`
	SeenCount     int64     `json:"seen_count"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
}

// DefaultPath returns the default leaderboard DB path.
func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "SenPaiScanner", "leaderboard.db")
	}
	return "senpaiscanner_leaderboard.db"
}

// Open opens or creates the leaderboard database at path.
func Open(path string) (*DB, error) {
	if path == "" {
		path = DefaultPath()
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create store directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}
	s := &DB{db: db}
	if err := s.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database connection.
func (s *DB) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *DB) init(ctx context.Context) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA busy_timeout=5000;`,
		`CREATE TABLE IF NOT EXISTS results (
			ip TEXT PRIMARY KEY,
			port INTEGER NOT NULL,
			probe_mode TEXT NOT NULL,
			avg_ms REAL NOT NULL,
			min_ms REAL NOT NULL,
			max_ms REAL NOT NULL,
			jitter_ms REAL NOT NULL,
			loss_pct REAL NOT NULL,
			throughput_bps REAL NOT NULL,
			speed_tested INTEGER NOT NULL,
			colo TEXT NOT NULL,
			tls_ok INTEGER NOT NULL,
			http_status INTEGER NOT NULL,
			score REAL NOT NULL,
			best_score REAL NOT NULL,
			seen_count INTEGER NOT NULL,
			first_seen TEXT NOT NULL,
			last_seen TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_results_score ON results(score DESC, throughput_bps DESC, avg_ms ASC);`,
		`CREATE INDEX IF NOT EXISTS idx_results_last_seen ON results(last_seen DESC);`,
		`CREATE TABLE IF NOT EXISTS checkpoints (
			name TEXT PRIMARY KEY,
			offset INTEGER NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("init store: %w", err)
		}
	}
	return nil
}

// SaveResult upserts the latest measurement for an IP.
func (s *DB) SaveResult(ctx context.Context, r *result.Result) error {
	if s == nil || s.db == nil || r == nil || r.IP == nil {
		return nil
	}
	seen := r.Timestamp
	if seen.IsZero() {
		seen = time.Now()
	}
	score := r.Score()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO results (
			ip, port, probe_mode, avg_ms, min_ms, max_ms, jitter_ms, loss_pct,
			throughput_bps, speed_tested, colo, tls_ok, http_status, score,
			best_score, seen_count, first_seen, last_seen
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET
			port=excluded.port,
			probe_mode=excluded.probe_mode,
			avg_ms=excluded.avg_ms,
			min_ms=excluded.min_ms,
			max_ms=excluded.max_ms,
			jitter_ms=excluded.jitter_ms,
			loss_pct=excluded.loss_pct,
			throughput_bps=excluded.throughput_bps,
			speed_tested=excluded.speed_tested,
			colo=excluded.colo,
			tls_ok=excluded.tls_ok,
			http_status=excluded.http_status,
			score=excluded.score,
			best_score=max(results.best_score, excluded.best_score),
			seen_count=results.seen_count + 1,
			last_seen=excluded.last_seen
		`,
		r.IP.String(),
		r.Port,
		r.ProbeMode,
		result.Ms(r.Avg()),
		result.Ms(r.Min()),
		result.Ms(r.Max()),
		result.Ms(r.Jitter()),
		r.Loss(),
		r.Throughput,
		boolInt(r.SpeedTested),
		r.Colo,
		boolInt(r.TLSOk),
		r.HTTPStatus,
		score,
		score,
		seen.Format(time.RFC3339Nano),
		seen.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("save result: %w", err)
	}
	return nil
}

// TopIPs returns the best currently healthy IPs for warm-start scans.
func (s *DB) TopIPs(ctx context.Context, limit int) ([]net.IP, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT ip
		FROM results
		WHERE score > 0
		ORDER BY score DESC, throughput_bps DESC, avg_ms ASC, last_seen DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query warm IPs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ips []net.IP
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if ip := net.ParseIP(raw); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips, rows.Err()
}

// TopRecords returns the current leaderboard.
func (s *DB) TopRecords(ctx context.Context, limit int) ([]Record, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT ip, port, probe_mode, avg_ms, min_ms, max_ms, jitter_ms, loss_pct,
		       throughput_bps, speed_tested, colo, tls_ok, http_status, score,
		       best_score, seen_count, first_seen, last_seen
		FROM results
		WHERE score > 0
		ORDER BY score DESC, throughput_bps DESC, avg_ms ASC, last_seen DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Record
	for rows.Next() {
		var rec Record
		var speedTested, tlsOK int
		var firstSeen, lastSeen string
		if err := rows.Scan(
			&rec.IP, &rec.Port, &rec.ProbeMode, &rec.AvgMs, &rec.MinMs,
			&rec.MaxMs, &rec.JitterMs, &rec.LossPct, &rec.ThroughputBPS,
			&speedTested, &rec.Colo, &tlsOK, &rec.HTTPStatus, &rec.Score,
			&rec.BestScore, &rec.SeenCount, &firstSeen, &lastSeen,
		); err != nil {
			return nil, err
		}
		rec.SpeedTested = speedTested != 0
		rec.TLSOk = tlsOK != 0
		rec.FirstSeen, _ = time.Parse(time.RFC3339Nano, firstSeen)
		rec.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeen)
		out = append(out, rec)
	}
	if out == nil {
		out = []Record{}
	}
	return out, rows.Err()
}

// SaveCheckpoint stores a named scan offset for resume-capable sweeps.
func (s *DB) SaveCheckpoint(ctx context.Context, name string, offset int64) error {
	if name == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO checkpoints(name, offset, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			offset=excluded.offset,
			updated_at=excluded.updated_at`,
		name, offset, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}
	return nil
}

// LoadCheckpoint returns a named scan offset, if it exists.
func (s *DB) LoadCheckpoint(ctx context.Context, name string) (int64, bool, error) {
	var offset int64
	err := s.db.QueryRowContext(ctx, `SELECT offset FROM checkpoints WHERE name = ?`, name).Scan(&offset)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("load checkpoint: %w", err)
	}
	return offset, true, nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
