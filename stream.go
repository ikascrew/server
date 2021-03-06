package server

import (
	"fmt"
	"log"

	"github.com/ikascrew/core"
	"github.com/ikascrew/server/config"

	"gocv.io/x/gocv"
)

type Stream struct {
	now_video core.Video
	now_value float64
	now_image gocv.Mat

	old_video core.Video
	old_value float64
	old_image gocv.Mat

	release_video core.Video

	used map[string]bool

	nextFlag bool
	prevFlag bool

	empty_image gocv.Mat
	real_image  gocv.Mat

	light float64
	wait  float64

	mode int
}

const SWITCH_VALUE = 200

const (
	SWITCH = 0
	LIGHT  = 1
	WAIT   = 2
)

func NewStream() (*Stream, error) {

	rtn := Stream{}

	rtn.now_value = 0
	rtn.old_value = 0

	rtn.now_video = nil
	rtn.old_video = nil
	rtn.release_video = nil

	rtn.used = make(map[string]bool)
	conf := config.Get()

	w := conf.Width
	h := conf.Height

	rtn.now_image = gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)
	rtn.old_image = gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)

	rtn.nextFlag = false
	rtn.prevFlag = false

	rtn.empty_image = gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)
	rtn.real_image = gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)

	rtn.light = 0

	rtn.wait = 0
	rtn.mode = SWITCH
	return &rtn, nil
}

func (s *Stream) Switch(v core.Video) error {

	if s.used[v.Source()] {
		return fmt.Errorf("until used video")
	}
	s.used[v.Source()] = true

	s.old_value = s.now_value
	s.now_value = 0

	wk := s.release_video
	if wk != nil {
		delete(s.used, wk.Source())
		defer wk.Release()
	}

	s.release_video = s.old_video
	s.old_video = s.now_video
	s.now_video = v

	return nil
}

func (s *Stream) Add(org gocv.Mat) *gocv.Mat {

	alpha := s.light / 200 * -1
	gocv.AddWeighted(s.empty_image, float64(alpha), org, float64(1.0-alpha), 0.0, &s.real_image)

	return &s.real_image
}

func (s *Stream) Get() (*gocv.Mat, error) {

	old, err := s.getOldImage()
	if err != nil {
		return nil, err
	}

	if old == nil {
		//log.Printf("old == nil")
		return s.now_video.Next()
	}

	if s.nextFlag {
		if s.now_value == SWITCH_VALUE {
			s.nextFlag = false
		} else if s.now_value < SWITCH_VALUE {
			s.now_value++
		} else {
			s.now_value--
		}
	}

	if s.prevFlag {
		if s.now_value == 0 {
			s.prevFlag = false
		} else if s.now_value > 0 {
			s.now_value--
		} else {
			s.now_value++
		}
	}

	alpha := s.now_value / SWITCH_VALUE

	next, err := s.now_video.Next()
	if err != nil {
		log.Printf("Next video error", err)
		return nil, err
	}

	gocv.AddWeighted(*next, float64(alpha), *old, float64(1.0-alpha), 0.0, &s.now_image)

	return &s.now_image, nil
}

func (s *Stream) getOldImage() (*gocv.Mat, error) {

	if s.release_video == nil {
		if s.old_video != nil {
			return s.old_video.Next()
		}
		return nil, nil
	}

	alpha := s.old_value / SWITCH_VALUE

	next, _ := s.old_video.Next()
	if next == nil {
		return &s.old_image, nil
	}

	now, _ := s.release_video.Next()
	if now == nil {
		return &s.old_image, nil
	}

	gocv.AddWeighted(*next, float64(alpha), *now, float64(1.0-alpha), 0.0, &s.old_image)

	return &s.old_image, nil
}

func (s *Stream) Release() {

	s.now_image.Close()
	s.old_image.Close()

	if s.now_video != nil {
		s.now_video.Release()
	}
	if s.old_video != nil {
		s.old_video.Release()
	}
	if s.release_video != nil {
		s.release_video.Release()
	}
}

func (s *Stream) Wait() float64 {

	wait := 33.0 - (s.wait / 2)

	if wait <= 1.0 {
		wait = 1.0
	} else if wait >= 100.0 {
		wait = 100
	}

	return wait
}

func (s *Stream) PrintVideos(line string) {
	log.Printf(line + "-------------------------------------------------")
	if s.now_video != nil {
		log.Printf("[1]" + s.now_video.Source())
	}

	if s.old_video != nil {
		log.Printf("[2]" + s.old_video.Source())
	}

	if s.release_video != nil {
		log.Printf("[3]" + s.release_video.Source())
	}
}

func (s *Stream) SetSwitch(t string) error {
	if t == "next" {
		s.nextFlag = true
	} else if t == "prev" {
		s.prevFlag = true
	} else {
		return fmt.Errorf("Unknown type[%s]", t)
	}
	return nil
}
