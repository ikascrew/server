package server

import (
	"fmt"

	"github.com/ikascrew/core"
	"github.com/ikascrew/server/config"

	"gocv.io/x/gocv"
)

type Stream struct {
	videos *Videos

	nextFlag bool
	prevFlag bool

	empty_image gocv.Mat
	real_image  gocv.Mat

	light float64
	wait  float64

	mode int
}

type Videos struct {
	src []*VideoSet
}

func NewVideos(l int, w, h int) *Videos {
	var v Videos
	v.src = make([]*VideoSet, l)

	for idx := range v.src {
		v.src[idx] = new(VideoSet)
		v.src[idx].video = nil
		v.src[idx].value = 0.0
		v.src[idx].image = gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)
	}
	return &v
}

func (s *Videos) set(v core.Video) {

	last := len(s.src) - 1
	lastSet := s.v(last)
	defer lastSet.Release()

	for idx := last - 1; idx >= 0; idx-- {

		v1 := s.v(idx)
		v2 := s.v(idx + 1)

		v2.value = v1.value
		v2.video = v1.video
	}

	s.v(0).video = v
	s.v(0).value = 0.0
}

func (s *Videos) release() {
	for _, set := range s.src {
		set.Release()
	}
}

func (s *Videos) v(idx int) *VideoSet {
	return s.src[idx]
}

func (s *Videos) NowValue() float64 {
	return s.v(0).value
}

func (s *Videos) Get(val float64) (*gocv.Mat, error) {
	s.v(0).value = val
	return s.get(0)
}

func (s *Videos) get(idx int) (*gocv.Mat, error) {

	if idx >= len(s.src) {
		return nil, nil
	}

	v1 := s.v(idx)
	if v1 == nil || v1.video == nil {
		return nil, nil
	}
	next, _ := v1.video.Next()

	v2, _ := s.get(idx + 1)
	if v2 == nil {
		return next, nil
	}

	alpha := v1.value / SWITCH_VALUE

	gocv.AddWeighted(*next, float64(alpha), *v2, float64(1.0-alpha), 0.0, &v1.image)

	return &v1.image, nil
}

type VideoSet struct {
	video core.Video
	value float64
	image gocv.Mat
}

func (s *VideoSet) Release() {
	s.image.Close()
	if s.video != nil {
		s.video.Release()
		s.video = nil
	}
}

const SWITCH_VALUE = 200

const (
	SWITCH = 0
	LIGHT  = 1
	WAIT   = 2
)

func NewStream() (*Stream, error) {

	conf := config.Get()
	w := conf.Width
	h := conf.Height

	rtn := Stream{}
	rtn.videos = NewVideos(5, w, h)

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
	s.videos.set(v)
	return nil
}

func (s *Stream) Add(org gocv.Mat) *gocv.Mat {
	alpha := s.light / 200 * -1
	gocv.AddWeighted(s.empty_image, float64(alpha), org, float64(1.0-alpha), 0.0, &s.real_image)
	return &s.real_image
}

func (s *Stream) Get() (*gocv.Mat, error) {

	val := s.videos.NowValue()
	if s.nextFlag {
		if val == SWITCH_VALUE {
			s.nextFlag = false
		} else if val < SWITCH_VALUE {
			val++
		} else {
			val--
		}
	}

	if s.prevFlag {
		if val == 0 {
			s.prevFlag = false
		} else if val > 0 {
			val--
		} else {
			val++
		}
	}

	v, err := s.videos.Get(val)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Stream) Release() {
	s.videos.release()
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
