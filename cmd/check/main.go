package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/ikascrew/server"
	"github.com/ikascrew/server/config"

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

	//server.Start はウィンドウを閉じるまでブロックするため goroutine で起動し、
	//5秒以内に返ってきた場合は起動失敗としてエラーを持ち帰る
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	log.Println("Wait... 5 second")
	select {
	case err := <-startErr:
		return xerrors.Errorf("server start (run \"ika-server create <project-id>\" first?): %w", err)
	case <-time.After(5 * time.Second):
	}

	conf := config.Get()
	if conf == nil || len(conf.Contents) == 0 {
		return xerrors.New("no contents in work file. run \"ika-server create <project-id>\" first")
	}

	for key, data := range conf.Contents {
		log.Println(data)
		err := server.Set(key)
		if err != nil {
			log.Println(err)
		}

		limit := 5 * time.Second
		begin := time.Now()
		idx := 1
		for now := range time.Tick(1 * time.Second) {
			err = server.Put(idx)
			if err != nil {
				log.Println(err)
			}
			idx++
			if now.Sub(begin) >= limit {
				break
			}
		}
	}

	return nil
}
