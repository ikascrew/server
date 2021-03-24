package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/ikascrew/server"
	"golang.org/x/xerrors"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := run()
	if err != nil {
		fmt.Printf("ika-server start error: %+v", err)
		os.Exit(1)
	}

	fmt.Println("Bye!")
	os.Exit(0)
}

func run() error {

	flag.Parse()
	args := flag.Args()

	l := len(args)
	if l < 1 {
		return xerrors.Errorf("ika-server start arguments project id required")
	}

	project := args[1]
	p, err := strconv.Atoi(project)
	if err != nil {
		return xerrors.Errorf("project id is int value(%s): %w", project, err)
	}

	err = server.Start(p)
	if err != nil {
		return xerrors.Errorf("server start: %w", err)
	}

	return nil
}
