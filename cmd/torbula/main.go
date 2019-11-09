package main

import (
	"fmt"
	"os"

	"github.com/isutare412/torbula"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [setting.ini]\n", os.Args[0])
		os.Exit(1)
	}

	var server *torbula.Server
	server, err := torbula.NewServer(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "on NewServer: %v\n", err)
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "server stopped: %v\n", err)
		os.Exit(1)
	}
}
