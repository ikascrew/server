package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ikascrew/server/config"
)

// setupTestConfig writes a minimal work file into a fresh temp directory,
// chdirs into it (config.WorkPath is a relative path) and loads it into the
// package-level config singleton via config.Set. The original working
// directory is restored on test cleanup.
func setupTestConfig(t *testing.T, width, height int) *config.Config {
	t.Helper()

	dir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore chdir: %v", err)
		}
	})

	if err := os.MkdirAll(".server", 0755); err != nil {
		t.Fatalf("mkdir work dir: %v", err)
	}

	conf := config.Config{
		Port:      55555,
		DBIP:      "localhost",
		DBPort:    5555,
		ProjectID: 1,
		Width:     width,
		Height:    height,
		Default:   config.Default{Type: "terminal", Name: "blank"},
		Contents: map[int]*config.Content{
			1: {ContentID: 1, Name: "a", Path: "a.mp4", Type: "file"},
		},
	}

	buf, err := json.MarshalIndent(&conf, "", "  ")
	if err != nil {
		t.Fatalf("marshal work file: %v", err)
	}

	if err := os.WriteFile(config.WorkPath(), buf, 0644); err != nil {
		t.Fatalf("write work file: %v", err)
	}

	if err := config.Set(); err != nil {
		t.Fatalf("config.Set: %v", err)
	}

	return config.Get()
}
