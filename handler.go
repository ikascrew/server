package server

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/ikascrew/pb"
	"github.com/ikascrew/server/config"

	"golang.org/x/net/context"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
}

func (i *IkascrewServer) startRPC() error {

	conf := config.Get()

	host := fmt.Sprintf(":%d", conf.Port)

	log.Println("Listen gRPC " + host)

	lis, err := net.Listen("tcp", host)
	if err != nil {
		return xerrors.Errorf("tcp listen port(%s): %w", host, err)
	}

	s := grpc.NewServer()
	pb.RegisterIkascrewServer(s, i)

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return xerrors.Errorf("start grpc server: %w", err)
	}
	return nil
}

func (i *IkascrewServer) Sync(ctx context.Context, r *pb.SyncRequest) (*pb.SyncReply, error) {

	i.window.FullScreen()

	rep := &pb.SyncReply{
		Source: 0,
		Type:   "file",
	}

	return rep, nil
}

func (i *IkascrewServer) Effect(ctx context.Context, r *pb.EffectRequest) (*pb.EffectReply, error) {

	rep := &pb.EffectReply{
		Success: false,
	}

	conf := config.Get()

	content, ok := conf.Contents[int(r.Id)]
	if !ok {
		return nil, fmt.Errorf("Content not found[%d]", r.Id)
	}
	fmt.Printf("[%s]-[%s]\n", r.Type, content.Path)
	if strings.Index(content.Path, ".jpg") >= 0 ||
		strings.Index(content.Path, ".jpeg") >= 0 ||
		strings.Index(content.Path, ".png") >= 0 {
		r.Type = "img"
	}

	if strings.Index(content.Path, "jpg") != -1 ||
		strings.Index(content.Path, ".jpeg") != -1 ||
		strings.Index(content.Path, ".png") != -1 {
		r.Type = "img"
	}

	v, err := Get(r.Type, content.Path)
	if err != nil {
		return rep, err
	}

	err = i.window.Push(v)
	if err != nil {
		return nil, err
	}

	rep.Success = true
	return rep, nil
}

func (i *IkascrewServer) Switch(ctx context.Context, r *pb.SwitchRequest) (*pb.SwitchReply, error) {

	rep := &pb.SwitchReply{
		Success: false,
	}

	err := i.window.SetSwitch(r.Type)
	if err == nil {
		rep.Success = true
	}
	return rep, err
}

func (i *IkascrewServer) PutVolume(ctx context.Context, msg *pb.VolumeMessage) (*pb.VolumeReply, error) {

	rep := pb.VolumeReply{}

	idx := msg.Index
	val := msg.Value

	s := i.window.stream

	s.mode = int(idx)

	switch s.mode {
	case SWITCH:
		s.now_value = val
	case LIGHT:
		s.light = val
	case WAIT:
		s.wait = val
	}

	return &rep, nil

}
