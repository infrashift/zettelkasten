package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s, cleanup, err := newServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize server: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
