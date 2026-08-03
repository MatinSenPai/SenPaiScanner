package ui

import (
	"testing"
)

func TestLookupIranISP(t *testing.T) {
	tests := []struct {
		ip       string
		wantISP  string
		wantFind bool
	}{
		// Known IP in range: 37.156.155.200 (within 37.156.128.0 - 37.156.143.255)
		{
			ip:       "37.156.130.1",
			wantISP:  "Iran Telecommunication Company Pjs",
			wantFind: true,
		},
		// Known IP in range: 5.112.98.253 (within 5.112.0.0 - 5.127.255.255)
		{
			ip:       "5.112.98.253",
			wantISP:  "Iran Cell Service and Communication Company",
			wantFind: true,
		},
		// Known IP in range: 2.144.1.2 (within 2.144.0.0 - 2.147.255.255)
		{
			ip:       "2.144.1.2",
			wantISP:  "Iran Cell Service and Communication Company",
			wantFind: true,
		},
		// IP range boundary check (exact start of range)
		{
			ip:       "2.144.0.0",
			wantISP:  "Iran Cell Service and Communication Company",
			wantFind: true,
		},
		// IP range boundary check (exact end of range)
		{
			ip:       "2.147.255.255",
			wantISP:  "Iran Cell Service and Communication Company",
			wantFind: true,
		},
		// Non-Iranian IP: Cloudflare DNS
		{
			ip:       "1.1.1.1",
			wantISP:  "",
			wantFind: false,
		},
		// Non-Iranian IP: Google DNS
		{
			ip:       "8.8.8.8",
			wantISP:  "",
			wantFind: false,
		},
		// Invalid IP address
		{
			ip:       "not-an-ip",
			wantISP:  "",
			wantFind: false,
		},
		// Empty IP
		{
			ip:       "",
			wantISP:  "",
			wantFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			gotISP, gotFind := LookupIranISP(tt.ip)
			if gotFind != tt.wantFind {
				t.Errorf("LookupIranISP(%q) gotFind = %v, want %v", tt.ip, gotFind, tt.wantFind)
			}
			if gotISP != tt.wantISP {
				t.Errorf("LookupIranISP(%q) gotISP = %q, want %q", tt.ip, gotISP, tt.wantISP)
			}
		})
	}
}
