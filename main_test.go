package main

import "testing"

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name          string
		injected      string
		moduleVersion string
		want          string
	}{
		{name: "release ldflags", injected: "v1.1.1", moduleVersion: "(devel)", want: "v1.1.1"},
		{name: "go install module", injected: "dev", moduleVersion: "v1.1.1", want: "v1.1.1"},
		{name: "local build", injected: "dev", moduleVersion: "(devel)", want: "dev"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := versionFromBuildInfo(test.injected, test.moduleVersion); got != test.want {
				t.Fatalf("versionFromBuildInfo(%q, %q) = %q, want %q", test.injected, test.moduleVersion, got, test.want)
			}
		})
	}
}
