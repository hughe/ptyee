package ptyee

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// terminalState holds the terminal state for restoration on exit.
type terminalState struct {
	fd       int
	oldState *term.State
}

// setupRawMode sets the terminal to raw mode and returns a state object
// that can be used to restore the original state.
func setupRawMode(fd int) (*terminalState, error) {
	if !term.IsTerminal(fd) {
		return nil, nil // Not a terminal, nothing to do
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}

	return &terminalState{
		fd:       fd,
		oldState: oldState,
	}, nil
}

// restore restores the terminal to its original state.
func (ts *terminalState) restore() error {
	if ts == nil || ts.oldState == nil {
		return nil
	}
	return term.Restore(ts.fd, ts.oldState)
}

// setInitialWindowSize sets the initial PTY window size based on stdin's size.
func setInitialWindowSize(ptmx *os.File, stdin *os.File) error {
	ws, err := pty.GetsizeFull(stdin)
	if err != nil {
		return fmt.Errorf("failed to get window size: %w", err)
	}

	if err := pty.Setsize(ptmx, ws); err != nil {
		return fmt.Errorf("failed to set window size: %w", err)
	}

	return nil
}

// handleWindowSizeChanges monitors SIGWINCH signals and updates the PTY size accordingly.
// This runs until the context is cancelled.
func handleWindowSizeChanges(ctx context.Context, ptmx *os.File, stdin *os.File) {
	if !term.IsTerminal(int(stdin.Fd())) {
		return // Not a terminal, nothing to do
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			ws, err := pty.GetsizeFull(stdin)
			if err == nil {
				pty.Setsize(ptmx, ws)
			}
		}
	}
}
