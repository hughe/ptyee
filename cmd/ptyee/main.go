package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hugh/ptyee"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ptyee: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	socketPath := flag.String("socket", "/tmp/ptyee.socket", "Unix socket path for monitoring/injection")
	tagOutput := flag.Bool("tag-output", false, "Prefix each byte with 'T' (terminal) or 'P' (program)")
	jsonlOutput := flag.Bool("jsonl-output", false, "Output JSONL format (not yet implemented)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] -- COMMAND [ARGS...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --socket /tmp/my.sock -- bash\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --tag-output -- ls -la\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	// Check for conflicting flags
	if *tagOutput && *jsonlOutput {
		return fmt.Errorf("--tag-output and --jsonl-output are mutually exclusive")
	}

	// JSONL is not yet implemented
	if *jsonlOutput {
		return fmt.Errorf("--jsonl-output is not yet implemented")
	}

	// Get command arguments (everything after flags)
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return fmt.Errorf("no command specified")
	}

	// Determine output format
	format := ptyee.OutputRaw
	if *tagOutput {
		format = ptyee.OutputTagged
	}

	// Build configuration
	cfg := ptyee.Config{
		Command:      args,
		SocketPath:   *socketPath,
		OutputFormat: format,
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
	}

	// Create PTYTee instance
	pt, err := ptyee.NewPTYTee(cfg)
	if err != nil {
		return fmt.Errorf("failed to create PTYTee: %w", err)
	}

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Run PTYTee
	if err := pt.Run(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("run error: %w", err)
	}

	// Exit with the child process's exit code
	os.Exit(pt.ExitCode())
	return nil // unreachable
}
