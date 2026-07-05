package server

import (
	"log"
	"os"
	"os/signal"
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

	//Ctrl+C はデフォルトの即時終了を無効化し、プレイ中でない時のみ受け付ける
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	err := w.stream.Switch(v)
	if err != nil {
		return err
	}

	for {
		select {
		case v := <-w.wait:
			err := w.stream.Switch(v)
			if err != nil {
				log.Printf("Stream Push Error: %v", err)
			}
		case <-interrupt:
			if w.isPlaying() {
				log.Println("Interrupt ignored : now playing(fullscreen)")
			} else {
				log.Println("Interrupt : shutdown")
				w.Destroy()
				return nil
			}
		default:
			key, err := w.Display()
			if err != nil {
				log.Printf("Window Display Error: %v", err)
			}

			//プレイ中(フルスクリーン)は終了操作を受け付けない
			if !w.isPlaying() {
				if key == 27 { //ESC
					log.Println("ESC : shutdown")
					w.Destroy()
					return nil
				}
				if win.GetWindowProperty(gocv.WindowPropertyVisible) < 1 {
					log.Println("Window closed : shutdown")
					w.Destroy()
					return nil
				}
			}
		}
	}
}

func (w *Window) isPlaying() bool {
	return w.win.GetWindowProperty(gocv.WindowPropertyFullscreen) == float64(gocv.WindowFullscreen)
}

var counter = 0

func (w *Window) Display() (int, error) {

	//イメージを取得
	img, err := w.stream.Get()
	if err != nil {
		return -1, err
	}

	//作成
	add := w.stream.Add(*img)
	//表示
	w.win.IMShow(*add)
	key := w.win.WaitKey(int(w.stream.Wait()))

	return key, nil
}

func (w *Window) SetSwitch(t string) error {
	return w.stream.SetSwitch(t)
}

func (w *Window) Destroy() {
	w.stream.Release()
}

func (w *Window) FullScreen() {
	if w.win.GetWindowProperty(gocv.WindowPropertyFullscreen) == float64(gocv.WindowFullscreen) {
		w.win.SetWindowProperty(gocv.WindowPropertyFullscreen, gocv.WindowNormal)
	} else {
		w.win.SetWindowProperty(gocv.WindowPropertyFullscreen, gocv.WindowFullscreen)
	}
}
