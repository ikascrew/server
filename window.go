package server

import (
	"fmt"
	"log"
	"runtime"

	"github.com/ikascrew/core"

	"gocv.io/x/gocv"
)

func init() {
}

type Window struct {
	name string
	wait chan core.Video

	win *gocv.Window

	stream *Stream
}

func NewWindow(name string) (*Window, error) {

	rtn := &Window{}

	rtn.name = name
	rtn.wait = make(chan core.Video)

	var err error
	rtn.stream, err = NewStream()
	return rtn, err
}

func (w *Window) Push(v core.Video) error {
	w.wait <- v

	//w.stream.PrintVideos("Push")
	return nil
}

func (w *Window) Play(v core.Video) error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	win := gocv.NewWindow(w.name)
	defer win.Close()

	w.win = win

	win.MoveWindow(0, 0)
	win.ResizeWindow(640, 360)

	err := w.stream.Switch(v)
	if err != nil {
		return err
	}

	for {
		select {
		case v := <-w.wait:
			err := w.stream.Switch(v)
			if err != nil {
				log.Printf("Stream Push Error:", err)
			}
		default:
			err := w.Display()
			if err != nil {
				log.Printf("Window Display Error:", err)
			}
		}
	}

	return fmt.Errorf("Error : Stream is nil")
}

var counter = 0

func (w *Window) Display() error {

	/*
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			time.Sleep(w.stream.Wait() * time.Millisecond)
			wg.Done()
		}()
	*/

	//イメージを取得
	img, err := w.stream.Get()
	if err != nil {
		return err
	}

	//作成
	add := w.stream.Add(*img)
	//表示
	w.win.IMShow(*add)
	w.win.WaitKey(int(w.stream.Wait()))

	//wg.Wait()

	return nil
}

func (w *Window) SetSwitch(t string) error {
	return w.stream.SetSwitch(t)
}

func (w *Window) Destroy() {
	w.stream.Release()
}

func (w *Window) FullScreen() {

	val := w.win.GetWindowProperty(gocv.WindowPropertyFullscreen)
	if int(val) == int(gocv.WindowFullscreen) {
		w.win.SetWindowProperty(gocv.WindowPropertyFullscreen, gocv.WindowNormal)
		//w.win.SetWindowProperty(gocv.WindowPropertyAutosize, gocv.PropertyAutosize)
		//w.win.SetWindowProperty(gocv.WindowPropertyAspectRatio, gocv.WindowKeepRatio)
	} else {
		w.win.SetWindowProperty(gocv.WindowPropertyFullscreen, gocv.WindowFullscreen)
	}
}
