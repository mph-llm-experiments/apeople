package main

import (
	"fmt"
	"os"

	"github.com/mph-llm-experiments/apeople/internal/cli"
	"github.com/mph-llm-experiments/apeople/internal/config"
)

var version = "0.2.0"

func main() {
	// Check for version flag early
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Printf("apeople v%s\n", version)
			os.Exit(0)
		}
	}

	// Load initial config (may be overridden by global flags)
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Run CLI
	if err := cli.Run(cfg, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
