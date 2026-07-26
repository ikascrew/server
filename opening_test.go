package server

import (
	"strings"
	"testing"
)

func TestCreateTerminalContainsExpectedSections(t *testing.T) {
	b := createTerminal()
	got := b.String()

	wantSubstrings := []string{
		"I am ikascrew.",
		"Today's system:",
		"DISPLAY:1280 x 720",
		"I am a ready.",
		"Let's get started!",
	}

	for _, want := range wantSubstrings {
		if !strings.Contains(got, want) {
			t.Errorf("createTerminal() output missing %q\nfull output:\n%s", want, got)
		}
	}
}
