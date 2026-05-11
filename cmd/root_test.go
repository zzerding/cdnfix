package cmd

import (
	"bytes"
	"testing"
)

func TestWantsVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "short", args: []string{"-v"}, want: true},
		{name: "long", args: []string{"--version"}, want: true},
		{name: "subcommand", args: []string{"refresh", "-v"}, want: false},
		{name: "help", args: []string{"--help"}, want: false},
		{name: "empty", args: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wantsVersion(tt.args); got != tt.want {
				t.Fatalf("wantsVersion(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestExecuteVersionOnly(t *testing.T) {
	oldVersion := version
	version = "v1.2.3"
	defer func() {
		version = oldVersion
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := execute([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("execute returned code %d", code)
	}
	if got := stdout.String(); got != "v1.2.3\n" {
		t.Fatalf("stdout = %q, want %q", got, "v1.2.3\n")
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}
