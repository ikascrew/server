package server

import (
	"fmt"
	"log"

	mc "github.com/ikascrew/core/multicast"
	"github.com/ikascrew/server/config"

	"golang.org/x/xerrors"
)

func init() {
}

type IkascrewServer struct {
	window *Window
}

func Start(p int, opts ...config.Option) error {

	err := config.Set(p, opts...)
	if err != nil {
		return xerrors.Errorf("config error: %w", err)
	}

	//server multicast
	go func() {
		err := startMulticast()
		if err != nil {
			log.Println("start multicast : %+v", err)
		}
	}()

	buf := createTerminal()
	v, err := Get("terminal", buf.String())
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

func startMulticast() error {

	udp, err := mc.NewServer(
		mc.ServerName("ikascrew server"),
	)

	if err != nil {
		return xerrors.Errorf("udp open error: %w", err)
	}

	err = udp.Dial()
	if err != nil {
		return xerrors.Errorf("udp dial error: %w", err)
	}

	return nil
}
