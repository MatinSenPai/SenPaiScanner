package prober

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNormalizeWSPath(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "/"},
		{name: "already absolute", in: "/ray", want: "/ray"},
		{name: "relative", in: "ray", want: "/ray"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeWSPath(tc.in); got != tc.want {
				t.Fatalf("normalizeWSPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestProbeWebSocketUsesConfiguredHostAndPath(t *testing.T) {
	cert, err := testCertificate()
	if err != nil {
		t.Fatal(err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	requestCh := make(chan *http.Request, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			return
		}
		requestCh <- req
		_, _ = conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"))
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("listener host %q is not an IP", host)
	}
	portNum, err := net.LookupPort("tcp", port)
	if err != nil {
		t.Fatal(err)
	}

	ok := probeWebSocket(context.Background(), ip, portNum, "example.com", "worker.example.com", "/vless", 3*time.Second)
	if !ok {
		t.Fatal("expected websocket probe to succeed against local TLS server")
	}

	select {
	case req := <-requestCh:
		if req.URL.Path != "/vless" {
			t.Fatalf("path = %q, want /vless", req.URL.Path)
		}
		if req.Host != "worker.example.com" {
			t.Fatalf("host = %q, want worker.example.com", req.Host)
		}
		if !strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
			t.Fatalf("Upgrade header = %q, want websocket", req.Header.Get("Upgrade"))
		}
	case <-time.After(time.Second):
		t.Fatal("server did not receive websocket request")
	}
}

func TestProbeWebSocketTimesOutDuringTLSHandshake(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	releaseConn := make(chan struct{})
	defer close(releaseConn)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		<-releaseConn
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("listener host %q is not an IP", host)
	}
	portNum, err := net.LookupPort("tcp", port)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	ok := probeWebSocket(context.Background(), ip, portNum, "example.com", "", "/", 150*time.Millisecond)
	elapsed := time.Since(start)
	if ok {
		t.Fatal("expected websocket probe to fail against a stalled TLS server")
	}
	if elapsed > time.Second {
		t.Fatalf("websocket probe took %s, want it bounded by the probe timeout", elapsed)
	}
}

func testCertificate() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"example.com"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return tls.X509KeyPair(certPEM, keyPEM)
}
