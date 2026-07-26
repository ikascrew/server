package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/ikascrew/ikasbox/db"
	"github.com/ikascrew/ikasbox/handler"
)

// withTempWorkDir chdirs into a fresh temp directory for the duration of the
// test (WorkPath()/workDir are relative paths) and resets the package-level
// config singleton on cleanup so state never leaks between tests.
func withTempWorkDir(t *testing.T) {
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
		gConf = nil
	})
}

func TestPortOption(t *testing.T) {
	conf := defaultConfig()
	if err := Port(9999)(conf); err != nil {
		t.Fatalf("Port option: %v", err)
	}
	if conf.Port != 9999 {
		t.Errorf("Port = %d, want 9999", conf.Port)
	}
}

func TestDatabaseOption(t *testing.T) {
	conf := defaultConfig()
	if err := Database("example.com", 1234)(conf); err != nil {
		t.Fatalf("Database option: %v", err)
	}
	if conf.DBIP != "example.com" || conf.DBPort != 1234 {
		t.Errorf("DBIP/DBPort = %s/%d, want example.com/1234", conf.DBIP, conf.DBPort)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	if c.Port != 55555 {
		t.Errorf("default Port = %d, want 55555", c.Port)
	}
	if c.DBIP != "localhost" || c.DBPort != 5555 {
		t.Errorf("default DBIP/DBPort = %s/%d, want localhost/5555", c.DBIP, c.DBPort)
	}
}

func TestSetMissingWorkFileErrors(t *testing.T) {
	withTempWorkDir(t)

	if err := Set(); err == nil {
		t.Fatalf("expected an error when the work file does not exist")
	}
	if Get() != nil {
		t.Errorf("Get() should stay nil after a failed Set()")
	}
}

func TestSetAppliesOptionsOverWorkFile(t *testing.T) {
	withTempWorkDir(t)

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("mkdir work dir: %v", err)
	}

	conf := defaultConfig()
	conf.ProjectID = 3
	buf, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(WorkPath(), buf, 0644); err != nil {
		t.Fatalf("write work file: %v", err)
	}

	if err := Set(Port(1234)); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got := Get()
	if got == nil {
		t.Fatalf("Get() returned nil after Set()")
	}
	if got.Port != 1234 {
		t.Errorf("Port = %d, want 1234 (option should override work file)", got.Port)
	}
	if got.ProjectID != 3 {
		t.Errorf("ProjectID = %d, want 3 (from work file)", got.ProjectID)
	}
}

// TestCreateAndSetRoundTrip fakes the ikasbox HTTP endpoint that Create hits,
// checks the work file it writes, then verifies Set reads it back correctly.
func TestCreateAndSetRoundTrip(t *testing.T) {
	withTempWorkDir(t)

	res := handler.ProjectResponse{
		Project: &db.Project{ID: 7, Name: "proj", Width: 1280, Height: 720},
		Contents: []*db.Content{
			{ID: 1, Name: "a", Type: "file", Path: "a.mp4"},
			{ID: 2, Name: "b", Type: "cd", Params: `{"target":"2027-01-01T00:00:00+09:00"}`},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(res); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}

	if err := Create(7, Database(u.Hostname(), port)); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := os.Stat(WorkPath()); err != nil {
		t.Fatalf("expected work file to exist: %v", err)
	}

	if err := Set(); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got := Get()
	if got.ProjectID != 7 {
		t.Errorf("ProjectID = %d, want 7", got.ProjectID)
	}
	if got.Width != 1280 || got.Height != 720 {
		t.Errorf("Width/Height = %d/%d, want 1280/720", got.Width, got.Height)
	}
	if got.Default.Type != "terminal" || got.Default.Name != "blank" {
		t.Errorf("Default = %+v, want {terminal blank}", got.Default)
	}
	if len(got.Contents) != 2 {
		t.Fatalf("len(Contents) = %d, want 2", len(got.Contents))
	}

	c1, ok := got.Contents[1]
	if !ok {
		t.Fatalf("Contents[1] missing")
	}
	if c1.Type != "file" || c1.Path != "a.mp4" {
		t.Errorf("Contents[1] = %+v, want Type=file Path=a.mp4", c1)
	}

	c2, ok := got.Contents[2]
	if !ok {
		t.Fatalf("Contents[2] missing")
	}
	if c2.Type != "cd" || c2.Params == "" {
		t.Errorf("Contents[2] = %+v, want Type=cd with non-empty Params", c2)
	}
}
