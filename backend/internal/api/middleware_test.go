package api

import "testing"

func TestOriginAllowed(t *testing.T) {
	patterns := []string{
		"https://gopackage-chronicle.anphse.workers.dev",
		"https://*-gopackage-chronicle.anphse.workers.dev",
		"https://*.example.com",
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "exact production origin", origin: "https://gopackage-chronicle.anphse.workers.dev", want: true},
		{name: "workers branch preview", origin: "https://feat-backend-ci-and-fly-deploy-gopackage-chronicle.anphse.workers.dev", want: true},
		{name: "workers version preview", origin: "https://8f3a2b1c-gopackage-chronicle.anphse.workers.dev", want: true},
		{name: "other worker on same account", origin: "https://other-app.anphse.workers.dev", want: false},
		{name: "hyphen wildcard needs a non-empty prefix", origin: "https://-gopackage-chronicle.anphse.workers.dev", want: false},
		{name: "hyphen wildcard does not span dots", origin: "https://evil.feat-gopackage-chronicle.anphse.workers.dev", want: false},
		{name: "lookalike suffix on another domain", origin: "https://feat-gopackage-chronicle.anphse.workers.dev.evil.com", want: false},
		{name: "http scheme rejected", origin: "http://feat-gopackage-chronicle.anphse.workers.dev", want: false},
		{name: "dot wildcard one label", origin: "https://abc123.example.com", want: true},
		{name: "dot wildcard not apex", origin: "https://example.com", want: false},
		{name: "dot wildcard not two labels", origin: "https://a.b.example.com", want: false},
		{name: "uppercase rejected", origin: "https://FEAT-gopackage-chronicle.anphse.workers.dev", want: false},
		{name: "empty origin", origin: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := originAllowed(patterns, tt.origin); got != tt.want {
				t.Errorf("originAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
