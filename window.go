package server

import (
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/ikascrew/core"
	"github.com/ikascrew/core/window"

	"golang.org/x/xerrors"
)

func init() {
}

type Window struct {
	name string
	wait chan core.Video
	done chan struct{}

	win *window.Window

	stream *Stream

	// 終了処理の先頭で呼ばれるフック(gRPC サーバの停止など)。
	// ストリーム解放より先に外部からの操作を止めるために使う
	onShutdown func()
}

func NewWindow(name string) (*Window, error) {

	rtn := &Window{}

	rtn.name = name
	rtn.wait = make(chan core.Video)
	rtn.done = make(chan struct{})

	var err error
	rtn.stream, err = NewStream()
	return rtn, err
}

func (w *Window) Push(v core.Video) error {
	//シャットダウン後はレンダーループが受け取らないため、ブロックせずエラーを返す
	select {
	case w.wait <- v:
		return nil
	case <-w.done:
		return xerrors.New("server is shutting down")
	}
}

func (w *Window) Play(v core.Video) error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	win, err := window.New(w.name)
	if err != nil {
		return xerrors.Errorf("window new: %w", err)
	}
	//破棄済みウィンドウの再 Close 対策(cv::Exception)は window 側が吸収する
	defer win.Close()

	w.win = win

	win.Move(0, 0)
	win.Resize(640, 360)

	//Ctrl+C はデフォルトの即時終了を無効化し、プレイ中でない時のみ受け付ける
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	err = w.stream.Switch(v)
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
				return w.shutdown()
			}
		default:
			//☓や Alt+F4 で破棄済みのウィンドウへの描画・問い合わせは
			//cv::Exception でプロセスごと落ちるため、先に生存確認する
			if w.closed() {
				log.Println("Window closed : shutdown")
				return w.shutdown()
			}

			key, err := w.Display()
			if err != nil {
				log.Printf("Window Display Error: %v", err)
			}

			//プレイ中(フルスクリーン)は終了操作を受け付けない
			if !w.isPlaying() {
				if key == 27 { //ESC
					log.Println("ESC : shutdown")
					return w.shutdown()
				}
			}
		}
	}
}

// shutdown は安全な終了手順を実行する。
// 新規 Push を止め、gRPC サーバ等を停止してから動画・フレームを解放する。
// ウィンドウ自体は Play の defer(win.Close)で閉じる
func (w *Window) shutdown() error {
	close(w.done)
	if w.onShutdown != nil {
		w.onShutdown()
	}
	w.Destroy()
	return nil
}

// closed はウィンドウが破棄済みかを返す
func (w *Window) closed() bool {
	return w.win.Closed()
}

func (w *Window) isPlaying() bool {
	return w.win.IsFullscreen()
}

var counter = 0

func (w *Window) Display() (int, error) {

	//イメージを取得
	img, err := w.stream.Get()
	if err != nil {
		return -1, err
	}

	//作成
	add := w.stream.Add(img)
	//表示
	w.win.Show(add)
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
	w.win.ToggleFullscreen()
}
