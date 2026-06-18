package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	dir := t.TempDir()
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr: %v", err)
	}
	defer stderr.Close()

	exitCode := run([]string{"-version"}, stdin, stdout, stderr)
	if exitCode != 0 {
		data, _ := io.ReadAll(stderr)
		t.Fatalf("expected exit code 0, got %d stderr=%q", exitCode, string(data))
	}

	if _, err := stdout.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek stdout: %v", err)
	}
	data, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if !strings.Contains(string(data), "stellar-tui 0.1.0") {
		t.Fatalf("stdout = %q", string(data))
	}
}
