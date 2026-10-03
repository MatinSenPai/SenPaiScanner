package plaincli

import (
	"bytes"
	"strings"
	"testing"
)

func run(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := Run(args, strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunRejectsBadInput(t *testing.T) {
	cases := map[string][]string{
		"phase2-only without targets": {"scan", "-phase2-only", "-no-state"},
		"bad port":                    {"scan", "-ports", "abc", "-no-state"},
		"port out of range":           {"scan", "-ports", "70000", "-no-state"},
		"resume without state":        {"scan", "-resume", "-no-state"},
		"missing targets file":        {"scan", "-targets-file", "/definitely/not/here.txt", "-no-state"},
		"unknown flag":                {"scan", "-nope"},
	}
	for name, args := range cases {
		if code, _, _ := run(args...); code != 2 {
			t.Errorf("%s: exit code %d, want 2", name, code)
		}
	}
}

func TestRunHelpIsPlainText(t *testing.T) {
	code, _, errText := run("scan", "-h")
	if code != 0 || !strings.Contains(errText, "plain text mode") || strings.Contains(errText, "\x1b[") {
		t.Errorf("help must exit 0 and be plain text, got code=%d %q", code, errText)
	}
}

func TestOutputHasNoAnsiOrCarriageReturns(t *testing.T) {
	var out bytes.Buffer
	p := &printer{out: &out}
	p.Info("hello")
	p.Error("bad")
	p.Stats(5, 2, 10)
	p.progress()
	p.progress() // unchanged state must not be announced twice
	text := out.String()
	if strings.ContainsAny(text, "\r\x1b") {
		t.Errorf("screen reader output must not contain control codes: %q", text)
	}
	if strings.Count(text, "Progress:") != 1 {
		t.Errorf("an unchanged progress line must be announced once, got %q", text)
	}
}
