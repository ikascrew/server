package server

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

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
