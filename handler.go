package server

import (
	"fmt"
	"net"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ikascrew/pb"
	"github.com/ikascrew/server/config"
)

func init() {
}

func (i *IkascrewServer) startRPC() error {

	lis, err := net.Listen("tcp", ":55555")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterIkascrewServer(s, i)

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		fmt.Println("failed to serve: %v", err)
		panic(err)
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
		return nil, fmt.Errorf("Content not found[%dj]", r.Id)
	}

	fmt.Printf("[%s]-[%s]\n", r.Type, content.Path)

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