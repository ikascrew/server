package server

import (
	"context"
	"net"
	"testing"

	"github.com/ikascrew/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestIsImagePath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"a.jpg", true},
		{"A.JPG", true},
		{"path/to/b.jpeg", true},
		{"c.PNG", true},
		{"d.mp4", false},
		{"noext", false},
		{"", false},
		{"tricky.jpg.mp4", false},
	}

	for _, c := range cases {
		if got := isImagePath(c.path); got != c.want {
			t.Errorf("isImagePath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// bufDialer builds a grpc.WithContextDialer func backed by an in-memory
// bufconn listener, so the gRPC round trip is exercised without any real
// network socket.
func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

// newBufconnClient wires an IkascrewServer up to a real *grpc.Server served
// over bufconn, and returns a connected pb.IkascrewClient. The server and
// connection are both torn down via t.Cleanup.
func newBufconnClient(t *testing.T, ika *IkascrewServer) pb.IkascrewClient {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	ika.rpc = grpc.NewServer()
	pb.RegisterIkascrewServer(ika.rpc, ika)

	go func() {
		_ = ika.rpc.Serve(lis)
	}()
	t.Cleanup(ika.rpc.Stop)

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithInsecure())
	if err != nil {
		t.Fatalf("bufconn dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return pb.NewIkascrewClient(conn)
}

// TestPutVolumeGRPC drives PutVolume through an actual gRPC round trip
// (bufconn), covering the SWITCH/LIGHT/WAIT index dispatch. It intentionally
// never calls Window.Play, so no OpenCV window is ever created.
func TestPutVolumeGRPC(t *testing.T) {
	setupTestConfig(t, 64, 64)

	win, err := NewWindow("test-putvolume")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	defer win.Destroy()

	ika := &IkascrewServer{window: win}
	client := newBufconnClient(t, ika)

	cases := []struct {
		name     string
		index    int64
		value    float64
		wantMode int
	}{
		{"switch", SWITCH, 123.0, SWITCH},
		{"light", LIGHT, 50.0, LIGHT},
		{"wait", WAIT, 10.0, WAIT},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rep, err := client.PutVolume(context.Background(), &pb.VolumeMessage{Index: c.index, Value: c.value})
			if err != nil {
				t.Fatalf("PutVolume: %v", err)
			}
			if !rep.Success {
				t.Errorf("expected Success=true")
			}
			if len(rep.ValueList) != 3 {
				t.Fatalf("expected 3 values, got %d", len(rep.ValueList))
			}
			if win.stream.mode != c.wantMode {
				t.Errorf("stream.mode = %d, want %d", win.stream.mode, c.wantMode)
			}
		})
	}
}

// TestPutVolumeInvalidIndexPreservesValues checks the "read current value"
// behaviour: an out-of-range Index must not mutate the stream and must
// return the existing values (used by clients to resync on reconnect).
func TestPutVolumeInvalidIndexPreservesValues(t *testing.T) {
	setupTestConfig(t, 64, 64)

	win, err := NewWindow("test-putvolume-readonly")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	defer win.Destroy()

	win.stream.now_value = 42
	win.stream.light = 7
	win.stream.wait = 3
	win.stream.mode = SWITCH

	ika := &IkascrewServer{window: win}

	rep, err := ika.PutVolume(context.Background(), &pb.VolumeMessage{Index: 99, Value: 500})
	if err != nil {
		t.Fatalf("PutVolume: %v", err)
	}
	if !rep.Success {
		t.Errorf("expected Success=true")
	}

	want := []float64{42, 7, 3}
	if len(rep.ValueList) != len(want) {
		t.Fatalf("ValueList = %v, want %v", rep.ValueList, want)
	}
	for i := range want {
		if rep.ValueList[i] != want[i] {
			t.Errorf("ValueList[%d] = %v, want %v", i, rep.ValueList[i], want[i])
		}
	}
	if win.stream.mode != SWITCH {
		t.Errorf("mode changed unexpectedly to %d", win.stream.mode)
	}
}

// TestSwitchHandler exercises Switch's next/prev/invalid dispatch directly
// (no rendering involved: SetSwitch only flips a bool on the Stream).
func TestSwitchHandler(t *testing.T) {
	setupTestConfig(t, 8, 8)

	win, err := NewWindow("test-switch")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	defer win.Destroy()

	ika := &IkascrewServer{window: win}

	rep, err := ika.Switch(context.Background(), &pb.SwitchRequest{Type: "next"})
	if err != nil {
		t.Fatalf("Switch(next): %v", err)
	}
	if !rep.Success || !win.stream.nextFlag {
		t.Errorf("expected success and nextFlag=true, got success=%v nextFlag=%v", rep.Success, win.stream.nextFlag)
	}

	rep, err = ika.Switch(context.Background(), &pb.SwitchRequest{Type: "bogus"})
	if err == nil {
		t.Errorf("expected error for unknown switch type")
	}
	if rep.Success {
		t.Errorf("expected Success=false for unknown switch type")
	}
}
