package server

import (
	"errors"
	"testing"

	"github.com/ikascrew/core"
)

// fakeVideo is a minimal core.Video implementation for exercising Stream
// logic without decoding a real file.
type fakeVideo struct {
	source   string
	frame    *core.Frame
	nextErr  error
	released bool
}

func newFakeVideo(t *testing.T, source string, w, h int) *fakeVideo {
	t.Helper()
	f := &fakeVideo{
		source: source,
		frame:  core.NewFrame(w, h),
	}
	t.Cleanup(func() {
		if !f.released {
			f.frame.Close()
		}
	})
	return f
}

func (f *fakeVideo) Next() (*core.Frame, error) {
	if f.nextErr != nil {
		return nil, f.nextErr
	}
	return f.frame, nil
}

func (f *fakeVideo) Wait() float64  { return 33 }
func (f *fakeVideo) Set(int)        {}
func (f *fakeVideo) Current() int   { return 0 }
func (f *fakeVideo) Source() string { return f.source }
func (f *fakeVideo) Release() error {
	f.released = true
	f.frame.Close()
	return nil
}

func TestNewStreamDefaults(t *testing.T) {
	setupTestConfig(t, 8, 8)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	if s.now_value != 0 || s.old_value != 0 {
		t.Errorf("expected zero now/old value, got now=%v old=%v", s.now_value, s.old_value)
	}
	if s.mode != SWITCH {
		t.Errorf("expected default mode SWITCH, got %d", s.mode)
	}
	if s.light != 0 {
		t.Errorf("expected light=0, got %v", s.light)
	}
	if s.now_video != nil || s.old_video != nil || s.release_video != nil {
		t.Errorf("expected all video slots nil initially")
	}
}

// TestStreamSwitchRotation verifies the now/old/release rotation and that a
// video is only Released once it has fallen out of all three slots.
func TestStreamSwitchRotation(t *testing.T) {
	setupTestConfig(t, 8, 8)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	a := newFakeVideo(t, "a", 8, 8)
	b := newFakeVideo(t, "b", 8, 8)
	c := newFakeVideo(t, "c", 8, 8)
	d := newFakeVideo(t, "d", 8, 8)

	if err := s.Switch(a); err != nil {
		t.Fatalf("Switch(a): %v", err)
	}
	if s.now_video.Source() != "a" || s.old_video != nil || s.release_video != nil {
		t.Fatalf("unexpected state after 1st switch: now=%v old=%v release=%v",
			s.now_video, s.old_video, s.release_video)
	}

	if err := s.Switch(b); err != nil {
		t.Fatalf("Switch(b): %v", err)
	}
	if s.now_video.Source() != "b" || s.old_video.Source() != "a" || s.release_video != nil {
		t.Fatalf("unexpected state after 2nd switch: now=%s old=%s release=%v",
			s.now_video.Source(), s.old_video.Source(), s.release_video)
	}

	if err := s.Switch(c); err != nil {
		t.Fatalf("Switch(c): %v", err)
	}
	if s.now_video.Source() != "c" || s.old_video.Source() != "b" || s.release_video.Source() != "a" {
		t.Fatalf("unexpected state after 3rd switch")
	}
	if a.released {
		t.Errorf("a should not be released yet (still in release slot)")
	}

	if err := s.Switch(d); err != nil {
		t.Fatalf("Switch(d): %v", err)
	}
	if !a.released {
		t.Errorf("a should be released once it falls out of all three slots")
	}
	if s.now_video.Source() != "d" || s.old_video.Source() != "c" || s.release_video.Source() != "b" {
		t.Fatalf("unexpected state after 4th switch")
	}
}

// TestStreamSwitchDuplicateSourceAllowed documents that pushing a video whose
// Source() matches one already held is accepted (logged only), not an error.
func TestStreamSwitchDuplicateSourceAllowed(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	v1 := newFakeVideo(t, "dup", 4, 4)
	v2 := newFakeVideo(t, "dup", 4, 4)

	if err := s.Switch(v1); err != nil {
		t.Fatalf("Switch(v1): %v", err)
	}
	if err := s.Switch(v2); err != nil {
		t.Fatalf("duplicate source should not error: %v", err)
	}
}

func TestStreamAddNoLight(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	img := core.NewFrame(4, 4)
	defer img.Close()

	out := s.Add(img)
	if out.Rows() != 4 || out.Cols() != 4 {
		t.Errorf("unexpected output size %dx%d", out.Rows(), out.Cols())
	}
}

// TestStreamAddWithLight pins down Add's blend formula against the empty
// (all-zero) image: real = empty*alpha + org*(1-alpha), alpha = light/200*-1.
// For light=100 that's alpha=-0.5, so a mid-gray source (100) should come out
// at 100*1.5=150.
func TestStreamAddWithLight(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	s.light = 100

	img := core.NewFrame(4, 4)
	defer img.Close()
	img.Fill(100, 100, 100)

	out := s.Add(img)
	if out.Rows() != 4 || out.Cols() != 4 {
		t.Errorf("unexpected output size %dx%d", out.Rows(), out.Cols())
	}

	const want = 150.0
	const tolerance = 1.0
	mean := out.Mean()
	if diff := mean[0] - want; diff > tolerance || diff < -tolerance {
		t.Errorf("out.Mean()[0] = %v, want ~%v", mean[0], want)
	}
}

func TestStreamWait(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	cases := []struct {
		wait float64
		want float64
	}{
		{0, 33},
		{66, 1},     // 33 - 33 = 0, clamped to 1
		{-134, 100}, // 33 + 67 = 100, exactly the upper clamp
		{-300, 100}, // 33 + 150 = 183, clamped to 100
		{200, 1},    // 33 - 100 = -67, clamped to 1
	}

	for _, c := range cases {
		s.wait = c.wait
		if got := s.Wait(); got != c.want {
			t.Errorf("Wait() with wait=%v = %v, want %v", c.wait, got, c.want)
		}
	}
}

func TestStreamSetSwitch(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	if err := s.SetSwitch("next"); err != nil || !s.nextFlag {
		t.Errorf("SetSwitch(next) failed: err=%v nextFlag=%v", err, s.nextFlag)
	}
	s.nextFlag = false

	if err := s.SetSwitch("prev"); err != nil || !s.prevFlag {
		t.Errorf("SetSwitch(prev) failed: err=%v prevFlag=%v", err, s.prevFlag)
	}
	s.prevFlag = false

	if err := s.SetSwitch("bogus"); err == nil {
		t.Errorf("expected error for unknown switch type")
	}
}

// TestStreamGetOnlyNowVideo covers the common case of a single active video
// (no crossfade in progress): Get must return that video's frame verbatim.
func TestStreamGetOnlyNowVideo(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	v := newFakeVideo(t, "solo", 4, 4)
	if err := s.Switch(v); err != nil {
		t.Fatalf("Switch: %v", err)
	}

	img, err := s.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if img != v.frame {
		t.Errorf("expected Get() to return the now_video frame directly")
	}
}

// TestStreamGetPropagatesNextError covers safeNext error propagation when
// there is nothing to fall back to (no old video).
func TestStreamGetPropagatesNextError(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	v := newFakeVideo(t, "erroring", 4, 4)
	v.nextErr = errors.New("boom")
	if err := s.Switch(v); err != nil {
		t.Fatalf("Switch: %v", err)
	}

	img, err := s.Get()
	if err == nil {
		t.Errorf("expected error from Get()")
	}
	if img != nil {
		t.Errorf("expected nil image on error, got %v", img)
	}
}

// TestStreamGetCrossfadeProgression drives the next-flag driven crossfade:
// now_value should ramp from 0 towards SWITCH_VALUE, one step per Get() call,
// and the flag should clear only once the target is reached.
func TestStreamGetCrossfadeProgression(t *testing.T) {
	setupTestConfig(t, 4, 4)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	oldV := newFakeVideo(t, "old", 4, 4)
	nowV := newFakeVideo(t, "now", 4, 4)

	if err := s.Switch(oldV); err != nil {
		t.Fatalf("Switch(oldV): %v", err)
	}
	if err := s.Switch(nowV); err != nil {
		t.Fatalf("Switch(nowV): %v", err)
	}

	s.now_value = SWITCH_VALUE - 1
	s.nextFlag = true

	if _, err := s.Get(); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.now_value != SWITCH_VALUE {
		t.Fatalf("expected now_value=%v after reaching the target, got %v", SWITCH_VALUE, s.now_value)
	}
	if !s.nextFlag {
		t.Errorf("nextFlag should still be set on the call that reaches the target")
	}

	if _, err := s.Get(); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.nextFlag {
		t.Errorf("nextFlag should clear once now_value has reached SWITCH_VALUE")
	}
	if s.now_value != SWITCH_VALUE {
		t.Errorf("now_value should not move past SWITCH_VALUE, got %v", s.now_value)
	}
}

// TestStreamGetFrameSizeMismatchIsSafe ensures a resolution mismatch between
// slots is handled by falling back to the now-frame (logged), not a crash or
// error, matching the documented safety behaviour in stream.go.
func TestStreamGetFrameSizeMismatchIsSafe(t *testing.T) {
	setupTestConfig(t, 8, 8)

	s, err := NewStream()
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer s.Release()

	oldV := newFakeVideo(t, "old-small", 4, 4)
	nowV := newFakeVideo(t, "now-big", 8, 8)

	if err := s.Switch(oldV); err != nil {
		t.Fatalf("Switch(oldV): %v", err)
	}
	if err := s.Switch(nowV); err != nil {
		t.Fatalf("Switch(nowV): %v", err)
	}

	s.now_value = 100 // midway alpha, would normally trigger AddWeighted

	img, err := s.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if img == nil {
		t.Fatalf("expected a non-nil frame on size mismatch fallback")
	}
	if img.Cols() != 8 || img.Rows() != 8 {
		t.Errorf("expected fallback to the now-frame (8x8), got %dx%d", img.Cols(), img.Rows())
	}
}
