package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ikascrew/server"
)

const VERSION = "0.1.0"

func main() {

	flag.Parse()
	args := flag.Args()

	l := len(args)
	fmt.Println(args)
	if l < 1 {
		os.Exit(1)
	}

	project := args[0]

	err := server.Start(project)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
