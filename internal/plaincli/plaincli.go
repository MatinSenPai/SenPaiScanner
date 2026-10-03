// Package plaincli is the screen-reader friendly front end: no full-screen UI, no colours, no redraws.
// Every message is a complete sentence on its own line, so NVDA / JAWS / Orca read it as it arrives.
package plaincli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/scanjob"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

const maxHitAnnouncements = 25

// Usage is shown for -h and for argument mistakes.
const Usage = `SenPai Scanner - plain text mode (no full-screen UI, made for screen readers and scripts)

Usage:
  senpaiscanner scan [options]        run a scan with the options below
  senpaiscanner --plain               answer a few questions, then scan
  senpaiscanner scan --resume         continue the scan that was interrupted

Options:
  -count N          how many random Cloudflare IPs to test (default 1000)
  -workers N        parallel probes (default 50)
  -timeout 5s       probe timeout (default 5s)
  -ports 443,8443   ports to test; 0 or empty uses the config's port, else 443
  -config URL       vless:// trojan:// vmess:// link; healthy IPs are then tested through it
  -top N            how many of the best IPs go to the tunnel test (default 50, 0 = all)
  -targets TEXT     test these instead of a random pool: IPs, CIDRs, ranges (a.b.c.d-e.f.g.h), domains
  -targets-file F   same, read from a file ("-" reads standard input)
  -phase2-only      skip the reachability scan and test the targets directly
                    (speed test through -config, or a direct download sample without it)
  -gentle           low-impact mode for ISPs that cut the connection during scans
                    (at most 25 workers, at least 6 s timeout, 40 probes per second)
  -ws               require a WebSocket answer
  -neighbors        also probe addresses next to healthy ones
  -min-speed Mbps   reject tunnel results slower than this
  -speed-url URL    download URL for the tunnel speed test
  -speed-size BYTES size of the speed sample (default 524288)
  -upload           also measure upload speed
  -antidpi=false    turn the ClientHello fragmentation off (on by default)
  -progress 15s     how often to announce progress (default 15s)
  -output FILE      also write the final healthy endpoints to this file
  -state FILE       where the resumable scan state is kept (default: your config folder)
  -no-state         do not keep resume data
  -resume           continue the interrupted scan (other options are taken from the saved scan)
`

type printer struct {
	mu       sync.Mutex
	out      io.Writer
	hits     int
	tested   int
	healthy  int
	total    int
	lastLine string // last progress sentence, so an unchanged state is not read out again
}

func (p *printer) line(format string, a ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.out, format+"\n", a...)
}

func (p *printer) Phase(phase int, livePath string) {
	if phase == 1 {
		p.line("Phase 1 started: checking which Cloudflare addresses answer.")
		if livePath != "" {
			p.line("Live results are written to %s", livePath)
		}
		return
	}
	p.line("Phase 2 started: testing the best addresses.")
}

func (p *printer) Stats(tested, healthy, total int) {
	p.mu.Lock()
	p.tested, p.healthy, p.total = tested, healthy, total
	p.mu.Unlock()
}

func (p *printer) Results(batch []*result.Result) {
	for _, r := range batch {
		if !r.IsHealthy() {
			continue
		}
		p.mu.Lock()
		p.hits++
		n := p.hits
		p.mu.Unlock()
		switch {
		case n <= maxHitAnnouncements:
			p.line("Healthy address found: %s, %d milliseconds, location %s, loss %.0f percent.",
				ui.FormatEndpoint(r.IP.String(), r.Port), r.Avg().Milliseconds(), orDash(r.Colo), r.Loss())
		case n == maxHitAnnouncements+1:
			p.line("More healthy addresses are being found. They will no longer be announced one by one; see the progress lines and the final list.")
		}
	}
}

func (p *printer) Validate(v *xraytest.ValidationResult, done, total int) {
	if v.Success {
		p.line("Test %d of %d passed: %s, download %s, latency %d milliseconds.", done, total,
			ui.FormatEndpoint(v.IP, v.Port), ui.FormatSpeed(v.Throughput), v.Latency.Milliseconds())
		return
	}
	reason := v.Error
	if reason == "" {
		reason = "no reason reported"
	}
	p.line("Test %d of %d failed: %s, %s.", done, total, ui.FormatEndpoint(v.IP, v.Port), reason)
}

func (p *printer) Info(msg string)  { p.line("%s", msg) }
func (p *printer) Error(msg string) { p.line("Problem: %s", msg) }

func (p *printer) progress() {
	p.mu.Lock()
	t, h, total := p.tested, p.healthy, p.total
	p.mu.Unlock()
	if total <= 0 {
		return
	}
	msg := fmt.Sprintf("Progress: %d of %d tested, %d percent, %d healthy.", t, total, t*100/total, h)
	p.mu.Lock()
	same := msg == p.lastLine
	p.lastLine = msg
	p.mu.Unlock()
	if !same {
		p.line("%s", msg)
	}
}

func orDash(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// Run executes the plain mode. args are the command line after the program name ("scan ..." or "--plain ...").
// It returns the process exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	interactive := len(args) > 0 && isPlainFlag(args[0])
	if len(args) > 0 && (args[0] == "scan" || isPlainFlag(args[0])) {
		args = args[1:]
	}
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, Usage) }
	var (
		count       = fs.Int("count", 1000, "")
		workers     = fs.Int("workers", 50, "")
		timeout     = fs.Duration("timeout", 5*time.Second, "")
		portsText   = fs.String("ports", "", "")
		configURL   = fs.String("config", "", "")
		top         = fs.Int("top", 50, "")
		targets     = fs.String("targets", "", "")
		targetsFile = fs.String("targets-file", "", "")
		phase2Only  = fs.Bool("phase2-only", false, "")
		gentle      = fs.Bool("gentle", false, "")
		requireWS   = fs.Bool("ws", false, "")
		neighbors   = fs.Bool("neighbors", false, "")
		minSpeed    = fs.Float64("min-speed", 0, "")
		speedURL    = fs.String("speed-url", "", "")
		speedSize   = fs.Int64("speed-size", 512*1024, "")
		upload      = fs.Bool("upload", false, "")
		antiDPI     = fs.Bool("antidpi", true, "")
		progress    = fs.Duration("progress", 15*time.Second, "")
		output      = fs.String("output", "", "")
		state       = fs.String("state", scanjob.DefaultStatePath(), "")
		noState     = fs.Bool("no-state", false, "")
		resume      = fs.Bool("resume", false, "")
	)
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	out := &printer{out: stdout}

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if interactive && len(set) == 0 {
		in := bufio.NewReader(stdin)
		ask := func(q, def string) string {
			fmt.Fprintf(stdout, "%s [%s]: ", q, def)
			s, _ := in.ReadString('\n')
			if s = strings.TrimSpace(s); s != "" {
				return s
			}
			return def
		}
		if info, ok := scanjob.InspectSnapshot(*state); ok &&
			strings.HasPrefix(strings.ToLower(ask(fmt.Sprintf("An interrupted scan from %s was found (%d of %d tested, %d healthy). Continue it? yes or no", info.Saved.Format("2006-01-02 15:04"), info.Tested, info.Total, info.Healthy), "yes")), "y") {
			*resume = true
		} else {
			*count, _ = strconv.Atoi(ask("How many random addresses should be tested", "1000"))
			*gentle = strings.HasPrefix(strings.ToLower(ask("Use gentle mode, for connections that drop during scans? yes or no", "no")), "y")
			*configURL = strings.TrimSpace(ask("Config link to test through (empty to skip)", ""))
			*targets = ask("Your own targets: IPs, ranges or domains separated by spaces (empty for a random pool)", "")
			if *targets != "" && *configURL != "" {
				*phase2Only = strings.HasPrefix(strings.ToLower(ask("Skip the reachability check and test them directly? yes or no", "no")), "y")
			}
		}
	}

	statePath := *state
	if *noState {
		statePath = ""
	}
	if *resume && statePath == "" {
		fmt.Fprintln(stderr, "Resuming needs resume data, but -no-state was given.")
		return 2
	}

	// anti-DPI: use the values saved by the desktop/terminal app, then apply the on/off flag
	if !*resume {
		ui.SetAntiDPI(ui.LoadAppConfig().LastConfig.AntiDPI)
		if set["antidpi"] {
			p := ui.CurrentAntiDPI()
			p.Enabled = *antiDPI
			ui.SetAntiDPI(p)
		}
	}

	text := *targets
	if *targetsFile != "" {
		var r io.Reader = stdin
		if *targetsFile != "-" {
			f, err := os.Open(*targetsFile)
			if err != nil {
				fmt.Fprintln(stderr, "Cannot open the targets file:", err)
				return 2
			}
			defer f.Close()
			r = f
		}
		b, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprintln(stderr, "Cannot read the targets:", err)
			return 2
		}
		text += "\n" + string(b)
	}
	if *phase2Only && strings.TrimSpace(text) == "" && !*resume {
		fmt.Fprintln(stderr, "-phase2-only needs -targets or -targets-file.")
		return 2
	}

	var ports []int
	for _, s := range strings.FieldsFunc(*portsText, func(r rune) bool { return r == ',' || r == ' ' }) {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 || n > 65535 {
			fmt.Fprintf(stderr, "%q is not a valid port.\n", s)
			return 2
		}
		ports = append(ports, n)
	}

	params := scanjob.Params{
		Count: *count, Workers: *workers, TimeoutMs: int(timeout.Milliseconds()), Ports: ports,
		ConfigURL: *configURL, RequireWS: *requireWS, TopN: *top, MinSpeed: *minSpeed, SpeedURL: *speedURL,
		SpeedSize: *speedSize, UploadTest: *upload, NeighborScan: *neighbors,
		Gentle: *gentle, Targets: text, Phase2Only: *phase2Only, StatePath: statePath, Resume: *resume,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	out.line("SenPai Scanner plain text mode. Press Control and C to stop; the scan can be continued later with the resume option.")
	if params.Gentle {
		out.line("Gentle mode is on: at most %d workers, at least %d seconds timeout, %.0f probes per second.", scanjob.GentleWorkers, int(scanjob.GentleTimeout.Seconds()), scanjob.GentleRate)
	}
	if p := ui.CurrentAntiDPI(); p.Enabled {
		out.line("Anti-DPI is on: the TLS handshake is sent in small pieces.")
	}

	ticker := time.NewTicker(*progress)
	defer ticker.Stop()
	finished := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				out.progress()
			case <-finished:
				return
			}
		}
	}()
	res := scanjob.Run(ctx, params, out)
	close(finished)
	out.progress()

	summarize(out, res, *output, statePath)
	return 0
}

func isPlainFlag(s string) bool {
	switch s {
	case "--plain", "--no-tui", "--simple", "--accessible":
		return true
	}
	return false
}

func summarize(out *printer, res scanjob.Outcome, outputPath, statePath string) {
	if res.Cancelled {
		out.line("The scan was stopped.")
		if statePath != "" {
			out.line("Its progress was saved. Run the scanner again with the resume option to continue where it stopped.")
		}
	} else {
		out.line("The scan finished.")
	}

	var lines []string
	if len(res.Working) > 0 {
		out.line("%d addresses passed the speed test:", len(res.Working))
		lines = res.Working
	} else if len(res.Phase1) > 0 {
		sorted := append([]*result.Result(nil), res.Phase1...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Avg() < sorted[j].Avg() })
		for _, r := range sorted {
			lines = append(lines, ui.FormatEndpoint(r.IP.String(), r.Port))
		}
		out.line("%d healthy addresses, fastest first:", len(lines))
	} else {
		out.line("No healthy address was found.")
	}
	for i, l := range lines {
		if i >= 20 {
			out.line("... and %d more (see the live results file).", len(lines)-20)
			break
		}
		out.line("%s", l)
	}
	if outputPath != "" && len(lines) > 0 {
		if err := os.WriteFile(outputPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			out.line("Problem: could not write %s: %v", outputPath, err)
		} else {
			out.line("All %d addresses were written to %s.", len(lines), outputPath)
		}
	}
}
