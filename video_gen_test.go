package server

import (
	"errors"
	"testing"

	"github.com/ikascrew/plugin/video"
)

// TestGetKnownType exercises the thin wrapper over the plugin/video registry
// for a generative type ("terminal") that needs no backing file.
func TestGetKnownType(t *testing.T) {
	v, err := Get("terminal", "hello")
	if err != nil {
		t.Fatalf("Get(terminal): %v", err)
	}
	if v == nil {
		t.Fatalf("expected non-nil video")
	}
	defer v.Release()

	if _, err := v.Next(); err != nil {
		t.Errorf("Next(): %v", err)
	}
}

// TestGetUnknownTypeWrapsNotFoundError checks that an unregistered type
// surfaces video.NotFoundError (server.NotFoundError is an alias of it) via
// errors.Is, so callers can distinguish "unknown type" from other failures.
func TestGetUnknownTypeWrapsNotFoundError(t *testing.T) {
	_, err := Get("does-not-exist", "")
	if err == nil {
		t.Fatalf("expected an error for an unknown type")
	}
	if !errors.Is(err, NotFoundError) {
		t.Errorf("expected error to wrap NotFoundError, got: %v", err)
	}
	if !errors.Is(err, video.NotFoundError) {
		t.Errorf("expected error to wrap video.NotFoundError, got: %v", err)
	}
}

// TestGetLegacyTypeNameNormalized checks that video.Normalize's legacy
// vocabulary absorption ("countdown" -> "cd") is reachable through server's
// Get wrapper, since Effect relies on this for old work files.
func TestGetLegacyTypeNameNormalized(t *testing.T) {
	v, err := Get("countdown", "")
	if err != nil {
		t.Fatalf("Get(countdown): %v", err)
	}
	defer v.Release()
}
