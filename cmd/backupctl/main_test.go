package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	cmd := newRootCommand()
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{nil, {"--help"}, {"backup", "--help"}, {"restore", "--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := execute(t, args...)
			if err != nil || !strings.Contains(output, "Usage:") {
				t.Fatalf("help output = %q, error = %v", output, err)
			}
		})
	}
}

func TestInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"unknown"}, {"backup", "extra"}, {"restore"},
		{"restore", "one", "two"}, {"backup", "--unknown"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := execute(t, args...); err == nil {
				t.Fatal("expected an argument error")
			}
		})
	}
}

func TestRestore(t *testing.T) {
	t.Chdir(t.TempDir())
	const snapshot = "snapshot-version=4\n"
	if err := os.WriteFile("snapshot.txt", []byte(snapshot), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := execute(t, "restore", "restored snapshot.txt")
	if err != nil {
		t.Fatalf("restore: %v\n%s", err, output)
	}
	data, err := os.ReadFile("restored snapshot.txt")
	if err != nil || string(data) != snapshot {
		t.Fatalf("restored data = %q, error = %v", data, err)
	}
	if !strings.Contains(output, "Command completed") {
		t.Fatalf("missing completion output: %q", output)
	}
	for _, destination := range []string{"snapshot.txt", "./snapshot.txt", filepath.Join(mustGetwd(t), "snapshot.txt")} {
		if _, err := execute(t, "restore", destination); err == nil || !strings.Contains(err.Error(), "must differ") {
			t.Fatalf("restore to %q: %v", destination, err)
		}
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestChildFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"backup"}, {"restore", "destination.txt"}} {
		output, err := execute(t, args...)
		if err == nil || !strings.Contains(err.Error(), "command failed") {
			t.Fatalf("%v: expected child failure, got %v", args, err)
		}
		if strings.Contains(output, "Command completed") || strings.Contains(output, "Usage:") {
			t.Fatalf("unexpected failure output: %q", output)
		}
	}
}

func TestBackup(t *testing.T) {
	t.Chdir(t.TempDir())
	// A stand-in publisher verifies subprocess execution and output forwarding.
	if err := os.WriteFile("mini-backup", []byte("#!/bin/sh\nprintf 'snapshot-data\\n' > snapshot.txt\nprintf 'published\\n'\nprintf 'diagnostic\\n' >&2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	output, err := execute(t, "backup")
	if err != nil {
		t.Fatalf("backup: %v\n%s", err, output)
	}
	data, err := os.ReadFile("snapshot.txt")
	if err != nil || string(data) != "snapshot-data\n" {
		t.Fatalf("snapshot = %q, error = %v", data, err)
	}
	for _, want := range []string{"published", "diagnostic", "Command completed"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}
