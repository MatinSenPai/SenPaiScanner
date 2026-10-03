package antidpi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The semantics below follow the PattN fork's fragment mask line by line
// (ported from Gemini Config Checker, internal/scan/fragment.go).

type fragCfg struct {
	hello      bool // packets: "tlshello"
	pFrom, pTo int64
	lenMin     []int64
	lenMax     []int64
	delMin     []int64
	delMax     []int64
	splitMin   int64
	splitMax   int64
}

// parseRange accepts 104, "104" or "100-200".
func parseRange(v any) (lo, hi int64, err error) {
	var s string
	switch x := v.(type) {
	case nil:
		return 0, 0, nil
	case float64:
		return int64(x), int64(x), nil
	case string:
		s = strings.TrimSpace(x)
	default:
		return 0, 0, fmt.Errorf("invalid value: %v", v)
	}
	if s == "" {
		return 0, 0, nil
	}
	a, b, found := strings.Cut(s, "-")
	if lo, err = strconv.ParseInt(strings.TrimSpace(a), 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid value: %q", s)
	}
	hi = lo
	if found {
		if hi, err = strconv.ParseInt(strings.TrimSpace(b), 10, 64); err != nil {
			return 0, 0, fmt.Errorf("invalid value: %q", s)
		}
	}
	if hi < lo {
		lo, hi = hi, lo
	}
	return lo, hi, nil
}

func rangeList(settings map[string]any, plural, single string) (mins, maxs []int64, err error) {
	if l, ok := settings[plural].([]any); ok && len(l) > 0 {
		for _, v := range l {
			lo, hi, e := parseRange(v)
			if e != nil {
				return nil, nil, e
			}
			mins, maxs = append(mins, lo), append(maxs, hi)
		}
		return
	}
	lo, hi, err := parseRange(settings[single])
	return []int64{lo}, []int64{hi}, err
}

// parseFinalmask reads {"tcp":[{"type":"fragment","settings":{...}}, ...]}. Only fragment masks are supported;
// anything else is reported instead of being silently ignored.
func parseFinalmask(text string) ([]fragCfg, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return nil, errors.New("finalmask: invalid JSON: " + err.Error())
	}
	list, _ := m["tcp"].([]any)
	var out []fragCfg
	for _, x := range list {
		mk, ok := x.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := mk["type"].(string); t != "fragment" {
			return nil, fmt.Errorf("finalmask: mask %q is not supported (only fragment)", mk["type"])
		}
		st, _ := mk["settings"].(map[string]any)
		if st == nil {
			st = map[string]any{}
		}
		var c fragCfg
		var err error
		switch p := st["packets"].(type) {
		case string:
			if strings.EqualFold(p, "tlshello") {
				c.hello, c.pFrom, c.pTo = true, 0, 1
			} else if c.pFrom, c.pTo, err = parseRange(p); err != nil {
				return nil, err
			}
		case float64:
			c.pFrom, c.pTo = int64(p), int64(p)
		}
		if !c.hello && c.pFrom == 0 {
			return nil, errors.New("finalmask: packets cannot be 0 or empty")
		}
		if c.lenMin, c.lenMax, err = rangeList(st, "lengths", "length"); err != nil {
			return nil, err
		}
		if c.lenMin[len(c.lenMin)-1] == 0 {
			return nil, errors.New("finalmask: the last value of lengths cannot be 0")
		}
		if c.delMin, c.delMax, err = rangeList(st, "delays", "delay"); err != nil {
			return nil, err
		}
		if c.splitMin, c.splitMax, err = parseRange(st["maxSplit"]); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func between(lo, hi int64) int64 {
	if hi <= lo {
		return lo
	}
	return lo + rand.Int63n(hi-lo+1)
}

// pick returns segment i's range, clamping to the last entry.
func pick(min, max []int64, i int) (int64, int64) {
	if i >= len(min) {
		i = len(min) - 1
	}
	return min[i], max[i]
}

type fragWriter struct {
	c     fragCfg
	w     io.Writer
	count uint64
}

func (f *fragWriter) Write(p []byte) (int, error) {
	f.count++
	c := f.c
	if c.hello {
		if f.count != 1 || len(p) <= 5 || p[0] != 22 {
			return f.w.Write(p)
		}
		recordLen := 5 + (int(p[3])<<8 | int(p[4]))
		if len(p) < recordLen {
			return f.w.Write(p)
		}
		data := p[5:recordLen]
		merge := len(c.delMax) == 1 && c.delMax[0] == 0 // all pieces leave in one write
		maxSplit := between(c.splitMin, c.splitMax)
		var merged []byte
		var n int64
		for from := 0; ; {
			lo, hi := pick(c.lenMin, c.lenMax, int(n))
			to := from + int(between(lo, hi))
			if to > len(data) || (maxSplit > 0 && n+1 >= maxSplit) {
				to = len(data)
			}
			l := to - from
			rec := make([]byte, 5+l)
			copy(rec[:3], p)
			rec[3], rec[4] = byte(l>>8), byte(l)
			copy(rec[5:], data[from:to])
			from = to
			if merge {
				merged = append(merged, rec...)
			} else {
				dlo, dhi := pick(c.delMin, c.delMax, int(n))
				if _, err := f.w.Write(rec); err != nil {
					return 0, err
				}
				if dhi > 0 {
					time.Sleep(time.Duration(between(dlo, dhi)) * time.Millisecond)
				}
			}
			n++
			if from == len(data) {
				if len(merged) > 0 {
					if _, err := f.w.Write(merged); err != nil {
						return 0, err
					}
				}
				if len(p) > recordLen {
					m, err := f.w.Write(p[recordLen:])
					if err != nil {
						return recordLen + m, err
					}
				}
				return len(p), nil
			}
		}
	}
	if c.pFrom != 0 && (f.count < uint64(c.pFrom) || f.count > uint64(c.pTo)) {
		return f.w.Write(p)
	}
	maxSplit := between(c.splitMin, c.splitMax)
	var n int64
	for from := 0; ; {
		lo, hi := pick(c.lenMin, c.lenMax, int(n))
		to := from + int(between(lo, hi))
		if to > len(p) || (maxSplit > 0 && n+1 >= maxSplit) {
			to = len(p)
		}
		m, err := f.w.Write(p[from:to])
		from += m
		if err != nil {
			return from, err
		}
		if dlo, dhi := pick(c.delMin, c.delMax, int(n)); dhi > 0 {
			time.Sleep(time.Duration(between(dlo, dhi)) * time.Millisecond)
		}
		n++
		if from >= len(p) {
			return from, nil
		}
	}
}

// chainWriters stacks the masks so that the first mask sees the application's writes and the last one writes to the wire.
func chainWriters(cfgs []fragCfg, wire io.Writer) io.Writer {
	w := wire
	for i := len(cfgs) - 1; i >= 0; i-- {
		w = &fragWriter{c: cfgs[i], w: w}
	}
	return w
}

// Proxy listens on 127.0.0.1 and forwards every connection to a target, fragmenting what goes out. It is how the
// embedded stock xray-core (which cannot parse the fork's mask schema) gets the same behaviour.
type Proxy struct {
	ln     net.Listener
	target string
	masks  []fragCfg
	once   sync.Once
}

// StartProxy returns nil, nil when the profile is disabled or has no masks (caller keeps the direct address).
func StartProxy(p Profile, target string) (*Proxy, error) {
	if !p.Enabled {
		return nil, nil
	}
	masks, err := parseFinalmask(p.Finalmask)
	if err != nil || len(masks) == 0 {
		return nil, err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	px := &Proxy{ln: ln, target: target, masks: masks}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go px.handle(c)
		}
	}()
	return px, nil
}

// Addr is the loopback address to dial instead of the target.
func (p *Proxy) Addr() (host string, port int) {
	return "127.0.0.1", p.ln.Addr().(*net.TCPAddr).Port
}

func (p *Proxy) Close() { p.once.Do(func() { p.ln.Close() }) }

func (p *Proxy) handle(c net.Conn) {
	defer c.Close()
	up, err := net.DialTimeout("tcp", p.target, 10*time.Second)
	if err != nil {
		return
	}
	defer up.Close()
	go func() { io.Copy(c, up); c.Close() }()

	w := chainWriters(p.masks, up)
	br := bufio.NewReaderSize(c, 64<<10)
	// Hand the first TLS record to the masks as ONE write, like a real TLS stack would.
	if h, err := br.Peek(5); err == nil && h[0] == 22 {
		rec := make([]byte, 5+(int(h[3])<<8|int(h[4])))
		if _, err := io.ReadFull(br, rec); err != nil {
			return
		}
		if _, err := w.Write(rec); err != nil {
			return
		}
	}
	buf := make([]byte, 32<<10)
	for {
		n, err := br.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}
