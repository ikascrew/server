package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/ikascrew/server"
	"github.com/ikascrew/server/config"
	"golang.org/x/xerrors"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := run()
	if err != nil {
		fmt.Printf("ika-server error: %+v", err)
		os.Exit(1)
	}

	fmt.Println("Bye!")
	os.Exit(0)
}

func run() error {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		return xerrors.New("usage: ika-server create <project-id> | ika-server start")
	}

	switch args[0] {
	case "create":
		return create(args[1:])
	case "start":
		return start()
	}

	return xerrors.Errorf("unknown command(%s)", args[0])
}

func create(args []string) error {

	if len(args) < 1 {
		return xerrors.New("create requires project id")
	}

	p, err := strconv.Atoi(args[0])
	if err != nil {
		return xerrors.Errorf("project id is int value(%s): %w", args[0], err)
	}

	err = config.Create(p)
	if err != nil {
		return xerrors.Errorf("create work: %w", err)
	}

	fmt.Printf("Created %s (project %d)\n", config.WorkPath(), p)
	return nil
}

func start() error {

	err := server.Start()
	if err != nil {
		return xerrors.Errorf("server start: %w", err)
	}

	return nil
}
