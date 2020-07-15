package server

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/ikascrew/server/config"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

func createTerminal() *strings.Builder {

	var b strings.Builder
	var err error

	cs, err := cpu.Info()
	cpuLine := make([]string, 0)
	if err == nil {
		c := cs[0]
		cpuLine = append(cpuLine, fmt.Sprintf("    CPU -> %s x %d x %d", c.ModelName, c.Cores, len(cs)))
	} else {
		cpuLine = append(cpuLine, fmt.Sprintf("    CPU Error :%s ", err.Error()))
	}

	memLine := make([]string, 0)
	m, err := mem.VirtualMemory()
	if err == nil {
		// structが返ってきます。
		memLine = append(memLine, fmt.Sprintf("    Mem:Total: %v, Free:%v", m.Total, m.Free))
	} else {
		memLine = append(memLine, fmt.Sprintf("    Mem Error :%s ", err.Error()))
	}

	dispLine := make([]string, 0)
	dispLine = append(dispLine, fmt.Sprintf("    DISPLAY:%d x %d", 1280, 720))

	//CPU
	//MEM
	b.WriteString("I am ikascrew.\n")
	b.WriteString("I am a program born to transform \"VJ System\".\n")
	b.WriteString("\n")
	b.WriteString("Today's system:\n")

	for _, line := range cpuLine {
		b.WriteString(line + "\n")
	}

	for _, line := range memLine {
		b.WriteString(line + "\n")
	}

	for _, line := range dispLine {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	b.WriteString("I am a ready.\n")
	b.WriteString("When you're ready?\n")
	b.WriteString("Let's get started!")

	return &b
}
