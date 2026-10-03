package prober

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
)

// A TLS probe must still succeed when its ClientHello is fragmented, and it must really arrive fragmented.
func TestProbeTLSWithAntiDPI(t *testing.T) {
	cert, err := testCertificate()
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	firstRead := make(chan int, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		// Peek the first segment on the wire, then hand the connection to the TLS server unchanged.
		buf := make([]byte, 4096)
		n, _ := c.(*net.TCPConn).Read(buf)
		firstRead <- n
		tc := tls.Server(&replayConn{Conn: c, first: buf[:n]}, &tls.Config{Certificates: []tls.Certificate{cert}})
		_ = tc.Handshake()
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(port)
	prof := antidpi.DefaultProfile()
	// Go's TLS server rejects the post's zero-length first record (Cloudflare accepts it); use a 1-byte one.
	prof.Finalmask = strings.Replace(prof.Finalmask, `["0", "104", "1"]`, `["1", "104", "1"]`, 1)
	ctx, err := antidpi.With(context.Background(), prof)
	if err != nil {
		t.Fatal(err)
	}
	lat, ok := probeTLS(ctx, net.ParseIP(host), p, "example.com", 5*time.Second, true)
	if !ok || lat <= 0 {
		t.Fatalf("probeTLS with Anti-DPI failed (ok=%v lat=%v)", ok, lat)
	}
	if n := <-firstRead; n > 120 {
		t.Errorf("first segment on the wire was %d bytes; the ClientHello was not fragmented", n)
	}
}

// replayConn re-delivers bytes that were already read before the TLS server took over.
type replayConn struct {
	net.Conn
	first []byte
}

func (c *replayConn) Read(p []byte) (int, error) {
	if len(c.first) > 0 {
		n := copy(p, c.first)
		c.first = c.first[n:]
		return n, nil
	}
	return c.Conn.Read(p)
}
