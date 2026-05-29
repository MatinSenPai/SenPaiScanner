package prober

// Cloudflare only proxies a fixed set of ports, so scanning anything else is
// wasted effort. These are Cloudflare's published HTTP/HTTPS proxied ports;
// keeping them in one place lets the UI offer a curated picker and lets the
// HTTP prober decide between cleartext and TLS without guessing.
//
// Source: Cloudflare "Network ports" documentation.
var (
	// HTTPSPorts terminate TLS (use mode tls or http).
	HTTPSPorts = []int{443, 2053, 2083, 2087, 2096, 8443}

	// HTTPPorts are cleartext HTTP (use mode http; no TLS handshake).
	HTTPPorts = []int{80, 8080, 8880, 2052, 2082, 2086, 2095}
)

// httpPortSet is a lookup built once from HTTPPorts.
var httpPortSet = func() map[int]struct{} {
	m := make(map[int]struct{}, len(HTTPPorts))
	for _, p := range HTTPPorts {
		m[p] = struct{}{}
	}
	return m
}()

// httpsPortSet is a lookup built once from HTTPSPorts.
var httpsPortSet = func() map[int]struct{} {
	m := make(map[int]struct{}, len(HTTPSPorts))
	for _, p := range HTTPSPorts {
		m[p] = struct{}{}
	}
	return m
}()

// IsCloudflarePort reports whether p is one of Cloudflare's proxied ports.
func IsCloudflarePort(p int) bool {
	if _, ok := httpsPortSet[p]; ok {
		return true
	}
	_, ok := httpPortSet[p]
	return ok
}

// isCleartextPort reports whether p should be spoken to over plain HTTP rather
// than HTTPS. Any known cleartext Cloudflare port qualifies; everything else
// (including unknown ports) defaults to TLS, which is the safe assumption for a
// 443-style edge.
func isCleartextPort(p int) bool {
	_, ok := httpPortSet[p]
	return ok
}
