package server

import (
	"context"
	"fmt"
	"log"

	mc "github.com/ikascrew/core/multicast"
	"github.com/ikascrew/pb"
	"github.com/ikascrew/server/config"

	"golang.org/x/xerrors"
)

func init() {
}

const Port = ":55555"

func Address() string {
	return Port
}

type IkascrewServer struct {
	window *Window
}

var server *IkascrewServer

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

	//start video
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

	server = ika

	return win.Play(v)
}

//test method
func Set(id int) error {
	req := pb.EffectRequest{}
	req.Id = int64(id)
	req.Type = "file"
	_, err := server.Effect(context.Background(), &req)
	return err
}

//test method
func Put(idx int) error {
	req := pb.VolumeMessage{}

	req.Index = int64(SWITCH)
	req.Value = float64(idx) / 5.0 * 200.0

	_, err := server.PutVolume(context.Background(), &req)

	return err
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
