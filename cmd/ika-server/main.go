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

var (
	ikasboxFlag = flag.Bool("ikasbox", false, "ikasbox(HTTP :5555)を同一プロセスで同居起動する")
	dbFlag      = flag.String("db", "ikasbox.db", "-ikasbox 時に使う ikasbox.db のパス")
)

func run() error {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		return xerrors.New("usage: ika-server create <project-id> | ika-server [-ikasbox [-db <path>]] start")
	}

	switch args[0] {
	case "create":
		return create(args[1:])
	case "start":
		return start(args[1:])
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

// start はサブコマンド後ろのフラグ("start -ikasbox ...")も受け付ける。
// Go 標準 flag は最初の非フラグ引数でパースを止めるため、
// グローバル側("-ikasbox start")の値を初期値にして再パースする
func start(args []string) error {

	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	ikasbox := fs.Bool("ikasbox", *ikasboxFlag, "ikasbox(HTTP :5555)を同一プロセスで同居起動する")
	db := fs.String("db", *dbFlag, "-ikasbox 時に使う ikasbox.db のパス")
	if err := fs.Parse(args); err != nil {
		return xerrors.Errorf("start flags: %w", err)
	}

	var opts []config.Option
	if *ikasbox {
		opts = append(opts, config.Ikasbox(*db))
	}

	err := server.Start(opts...)
	if err != nil {
		return xerrors.Errorf("server start: %w", err)
	}

	return nil
}
