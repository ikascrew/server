package server

import (
	"context"
	"fmt"
	"log"

	mc "github.com/ikascrew/core/multicast"
	"github.com/ikascrew/pb"
	"github.com/ikascrew/plugin/video/output"
	"github.com/ikascrew/server/config"

	"golang.org/x/xerrors"
)

func init() {
}

type IkascrewServer struct {
	// 前方互換のため生成コードの既定実装を埋め込む(protoc-gen-go-grpc の
	// mustEmbedUnimplementedIkascrewServer 要求)。.proto に RPC が増えても
	// この型は Unimplemented を返してコンパイルが通り続ける
	pb.UnimplementedIkascrewServer

	window *Window
}

var server *IkascrewServer

func Start(opts ...config.Option) error {

	err := config.Set(opts...)
	if err != nil {
		return xerrors.Errorf("config error: %w", err)
	}

	// 生成型プラグインにプロジェクトの解像度を伝える。
	// Stream はリサイズせずに合成するため、プラグインの描画サイズは
	// ここで設定した解像度と一致している必要がある。
	// 未作成(-ikasbox の初回起動)時は output 側の既定解像度に任せる
	conf := config.Get()
	if conf.Width > 0 && conf.Height > 0 {
		output.Set(conf.Width, conf.Height)
	}

	//server multicast
	go func() {
		err := startMulticast()
		if err != nil {
			log.Printf("start multicast : %+v", err)
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

	// ikasbox 同居モード(-ikasbox): HTTP :5555 を同一プロセスで起動し、
	// UI からの work file 作成(v1/server/create)を受け付ける
	if conf.Ikasbox {
		startIkasbox(conf.IkasboxDB)
	}

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
