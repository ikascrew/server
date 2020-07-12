package server

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/ikascrew/server/config"

	"golang.org/x/xerrors"
)

func init() {
}

const ADDRESS = ":55555"

func Address() string {
	return ADDRESS
}

type IkascrewServer struct {
	window *Window
}

func Start(d string) error {

	runtime.GOMAXPROCS(runtime.NumCPU())

	p, err := strconv.Atoi(d)
	if err != nil {
		return xerrors.Errorf("argument error: %w", err)
	}

	err = config.Load(p)
	if err != nil {
		return xerrors.Errorf("config error: %w", err)
	}

	//p64 := int64(p)
	//fmt.Println(p64)

	//TODO 初期化

	v, err := Get("terminal")
	if err != nil {
		return fmt.Errorf("Error:Video Load[%v]", err)
	}

	win, err := NewWindow("ikascrew")
	if err != nil {
		return fmt.Errorf("Error:Create New Window[%v]", err)
	}

	ika := &IkascrewServer{
		window: win,
	}

	go func() {
		ika.startRPC()
	}()

	return win.Play(v)
}
