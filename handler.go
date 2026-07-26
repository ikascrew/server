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

	i.rpc = grpc.NewServer()
	pb.RegisterIkascrewServer(i.rpc, i)

	reflection.Register(i.rpc)
	if err := i.rpc.Serve(lis); err != nil {
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

	// 型とパラメータは work file(ikasbox 由来)を正とし、
	// client からの型申告は Type を持たない旧 work file の救済にのみ使う。
	// 型名の正規化(image→img 等)は plugin/video 側で行われる
	t := content.Type
	if t == "" {
		t = r.Type
		if isImagePath(content.Path) {
			t = "img"
		}
	}

	param := content.Params
	if param == "" {
		param = content.Path
	}

	fmt.Printf("[%s]-[%s]\n", t, param)

	v, err := Get(t, param)
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

// isImagePath は旧 work file(Type 無し)向けの拡張子判定
func isImagePath(p string) bool {
	l := strings.ToLower(p)
	return strings.HasSuffix(l, ".jpg") ||
		strings.HasSuffix(l, ".jpeg") ||
		strings.HasSuffix(l, ".png")
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

// volumeList は現在の3値を Volume(SWITCH) / Light(LIGHT) / Wait(WAIT) の
// 順で返す。並び順は SWITCH=0 / LIGHT=1 / WAIT=2 の添字と一致させること
func (s *Stream) volumeList() []float64 {
	return []float64{s.now_value, s.light, s.wait}
}

func (i *IkascrewServer) PutVolume(ctx context.Context, msg *pb.VolumeMessage) (*pb.VolumeReply, error) {

	idx := int(msg.Index)
	val := msg.Value

	s := i.window.stream

	// 範囲外の Index では状態を一切変えない。壊れた値でモードを
	// 切り替えてしまうと本番中の描画に影響するため
	switch idx {
	case SWITCH:
		s.mode = idx
		s.now_value = val
	case LIGHT:
		s.mode = idx
		s.light = val
	case WAIT:
		s.mode = idx
		s.wait = val
	}

	return &pb.VolumeReply{Success: true, ValueList: s.volumeList()}, nil
}

// GetVolume は状態を変えずに現在値だけを返す。クライアントが接続し直した
// ときに手元の表示をサーバへ合わせるために使う
func (i *IkascrewServer) GetVolume(ctx context.Context, _ *pb.GetVolumeRequest) (*pb.VolumeReply, error) {

	s := i.window.stream

	return &pb.VolumeReply{Success: true, ValueList: s.volumeList()}, nil
}
