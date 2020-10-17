package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
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

	flag.Parse()
	args := flag.Args()

	l := len(args)
	if l < 1 {
		return xerrors.Errorf("ika-server start arguments project id required")
	}

	project := args[0]
	p, err := strconv.Atoi(project)
	if err != nil {
		return xerrors.Errorf("project id is int value(%s): %w", project, err)
	}

	go func() {
		err = server.Start(p)
		if err != nil {
			log.Printf("server start: %+v", err)
		}
	}()

	//10s
	log.Println("Wait... 5 second")
	time.Sleep(5 * time.Second)

	conf := config.Get()
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
