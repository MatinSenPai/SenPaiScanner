package antidpi

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type recorder struct{ writes [][]byte }

func (r *recorder) Write(p []byte) (int, error) {
	r.writes = append(r.writes, append([]byte(nil), p...))
	return len(p), nil
}

func fakeHello(n int) []byte {
	p := make([]byte, 5+n)
	p[0], p[1], p[2], p[3], p[4] = 22, 3, 1, byte(n>>8), byte(n)
	for i := 5; i < len(p); i++ {
		p[i] = byte(i)
	}
	return p
}

func postMask(t *testing.T) []fragCfg {
	t.Helper()
	cfgs, err := parseFinalmask(DefaultProfile().Finalmask)
	if err != nil || len(cfgs) != 2 {
		t.Fatalf("parse: %v (%d masks)", err, len(cfgs))
	}
	return cfgs
}

// The published values must cut a ClientHello byte for byte like the fork does.
func TestFragmentPostValues(t *testing.T) {
	hello := fakeHello(300)
	rec := &recorder{}
	if _, err := chainWriters(postMask(t), rec).Write(hello); err != nil {
		t.Fatal(err)
	}
	// mask 1 turns the hello into TLS records [0, 104, 1, 1, ...] sent as ONE write; mask 2 then cuts that write
	// into 114 bytes + single bytes, at most 11 writes.
	if len(rec.writes) != 11 || len(rec.writes[0]) != 114 || len(rec.writes[1]) != 1 || len(rec.writes[9]) != 1 {
		var sizes []int
		for _, w := range rec.writes {
			sizes = append(sizes, len(w))
		}
		t.Fatalf("TCP pieces wrong: %v", sizes)
	}
	var wire []byte
	for _, w := range rec.writes {
		wire = append(wire, w...)
	}
	var lens []int
	var payload []byte
	for len(wire) > 0 {
		l := int(wire[3])<<8 | int(wire[4])
		if wire[0] != 22 || wire[1] != 3 || wire[2] != 1 || len(wire) < 5+l {
			t.Fatalf("not a clean record stream at %d bytes left", len(wire))
		}
		lens, payload, wire = append(lens, l), append(payload, wire[5:5+l]...), wire[5+l:]
	}
	if lens[0] != 0 || lens[1] != 104 || lens[2] != 1 || len(lens) != 2+196 {
		t.Errorf("record sizes wrong: first=%v count=%d", lens[:3], len(lens))
	}
	if string(payload) != string(hello[5:]) {
		t.Error("the reassembled handshake differs from the original")
	}
}

func TestFragmentLeavesLaterTrafficAlone(t *testing.T) {
	rec := &recorder{}
	w := chainWriters(postMask(t), rec)
	w.Write(fakeHello(300))
	before := len(rec.writes)
	w.Write([]byte("application data that is long enough to be split if the mask were still active"))
	if len(rec.writes) != before+1 {
		t.Errorf("only the first write may be fragmented, got %d extra writes", len(rec.writes)-before)
	}
}

func TestValidate(t *testing.T) {
	if err := DefaultProfile().Validate(); err != nil {
		t.Fatalf("default profile must be valid: %v", err)
	}
	bad := DefaultProfile()
	bad.Finalmask = `{"tcp":[{"type":"noise"}]}`
	if bad.Validate() == nil {
		t.Error("unsupported mask type must be reported")
	}
	bad.Finalmask = `{not json`
	if bad.Validate() == nil {
		t.Error("invalid JSON must be reported")
	}
	bad.Enabled = false
	if bad.Validate() != nil {
		t.Error("a disabled profile is never invalid")
	}
}

// Go's strict TLS server rejects the post's zero-length first record (Cloudflare accepts it), so the handshake
// tests use the same recipe with a 1-byte first record: they prove a real stack completes through the fragmenter.
func strictProfile() Profile {
	p := DefaultProfile()
	p.Finalmask = strings.Replace(p.Finalmask, `["0", "104", "1"]`, `["1", "104", "1"]`, 1)
	return p
}

func TestProxyHandshake(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer srv.Close()
	px, err := StartProxy(strictProfile(), strings.TrimPrefix(srv.URL, "https://"))
	if err != nil || px == nil {
		t.Fatalf("proxy: %v", err)
	}
	defer px.Close()
	h, port := px.Addr()
	c, err := tls.Dial("tcp", net.JoinHostPort(h, strconv.Itoa(port)), &tls.Config{InsecureSkipVerify: true, ServerName: "example.com"})
	if err != nil {
		t.Fatalf("handshake through the proxy failed: %v", err)
	}
	c.Close()
}

func TestDialHandshake(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	ctx, err := With(context.Background(), strictProfile())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := Dial(ctx, &net.Dialer{}, "tcp", strings.TrimPrefix(srv.URL, "https://"))
	if err != nil {
		t.Fatal(err)
	}
	tc := tls.Client(conn, TLS(ctx, &tls.Config{InsecureSkipVerify: true, ServerName: "example.com"}))
	defer tc.Close()
	if err := tc.HandshakeContext(ctx); err != nil {
		t.Fatalf("handshake over the fragmenting conn failed: %v", err)
	}
	if got := tc.ConnectionState().NegotiatedProtocol; got != "" && got != "http/1.1" {
		t.Errorf("ALPN = %q, want http/1.1", got)
	}
}

func TestSuiteIDsSkipsTLS13(t *testing.T) {
	ids := suiteIDs(DefaultProfile().CipherSuites)
	if len(ids) != 10 { // 13 listed minus the 3 TLS 1.3 suites Go does not let you choose
		t.Errorf("got %d TLS 1.2 suites, want 10", len(ids))
	}
}
