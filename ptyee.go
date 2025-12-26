// Package ptyee provides functionality for teeing a PTY.
// It allows monitoring and injecting data into PTY sessions.
package ptyee

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/sync/errgroup"
)

// PTYTee manages a PTY session with socket-based monitoring and injection.
type PTYTee struct {
	cfg         Config
	cmd         *exec.Cmd
	ptmx        *os.File
	socketMgr   *socketManager
	termState   *terminalState
	exitCode    int
	exitCodeSet bool
}

// NewPTYTee creates a new PTYTee instance with the given configuration.
func NewPTYTee(cfg Config) (*PTYTee, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create socket manager
	socketMgr, err := newSocketManager(cfg.SocketPath, cfg.OutputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket manager: %w", err)
	}

	// Create command
	cmd := exec.Command(cfg.Command[0], cfg.Command[1:]...)

	return &PTYTee{
		cfg:       cfg,
		cmd:       cmd,
		socketMgr: socketMgr,
	}, nil
}

// Run executes the PTY session until the command completes or context is cancelled.
func (p *PTYTee) Run(ctx context.Context) error {
	// Start PTY
	ptmx, err := pty.Start(p.cmd)
	if err != nil {
		p.socketMgr.close()
		return fmt.Errorf("failed to start PTY: %w", err)
	}
	p.ptmx = ptmx
	defer p.cleanup()

	// Setup terminal raw mode if stdin is a terminal
	termState, err := setupRawMode(int(p.cfg.Stdin.(*os.File).Fd()))
	if err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}
	p.termState = termState

	// Set initial window size
	if stdin, ok := p.cfg.Stdin.(*os.File); ok {
		if err := setInitialWindowSize(ptmx, stdin); err != nil {
			// Non-fatal, just log and continue
			fmt.Fprintf(p.cfg.Stderr, "Warning: %v\n", err)
		}
	}

	// Use errgroup for coordinated goroutine management
	g, gctx := errgroup.WithContext(ctx)

	// Goroutine 1: Accept socket connections
	g.Go(func() error {
		return p.socketMgr.acceptConnection(gctx)
	})

	// Goroutine 2: Handle window size changes
	if stdin, ok := p.cfg.Stdin.(*os.File); ok {
		g.Go(func() error {
			handleWindowSizeChanges(gctx, ptmx, stdin)
			return nil
		})
	}

	// Goroutine 3: PTY output → stdout + socket
	g.Go(func() error {
		return p.copyPtyToOutputs(gctx)
	})

	// Goroutine 4: stdin → PTY
	g.Go(func() error {
		return p.copyStdinToPty(gctx)
	})

	// Goroutine 5: socket → PTY
	g.Go(func() error {
		return p.copySocketToPty(gctx)
	})

	// Goroutine 6: Wait for command completion
	g.Go(func() error {
		return p.waitForCommand(gctx)
	})

	// Wait for all goroutines (will complete when command finishes or context cancelled)
	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}

	return nil
}

// copyPtyToOutputs copies PTY output to stdout and the socket (byte by byte for tagging).
func (p *PTYTee) copyPtyToOutputs(ctx context.Context) error {
	buf := make([]byte, 1)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := p.ptmx.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("pty read error: %w", err)
		}

		if n > 0 {
			// Write to stdout
			if _, err := p.cfg.Stdout.Write(buf[:n]); err != nil {
				return fmt.Errorf("stdout write error: %w", err)
			}

			// Write to socket with 'P' tag (from program)
			if err := p.socketMgr.writeFromProgram(buf[0]); err != nil {
				// Socket write errors are non-fatal
				fmt.Fprintf(p.cfg.Stderr, "Warning: socket write error: %v\n", err)
			}
		}
	}
}

// copyStdinToPty copies stdin to PTY and also sends to socket with 'T' tag.
func (p *PTYTee) copyStdinToPty(ctx context.Context) error {
	buf := make([]byte, 1)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := p.cfg.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("stdin read error: %w", err)
		}

		if n > 0 {
			// Write to PTY
			if _, err := p.ptmx.Write(buf[:n]); err != nil {
				return fmt.Errorf("pty write error: %w", err)
			}

			// Write to socket with 'T' tag (from terminal)
			if err := p.socketMgr.writeFromTerminal(buf[0]); err != nil {
				// Socket write errors are non-fatal
				fmt.Fprintf(p.cfg.Stderr, "Warning: socket write error: %v\n", err)
			}
		}
	}
}

// copySocketToPty copies data from the socket to the PTY (for injection).
func (p *PTYTee) copySocketToPty(ctx context.Context) error {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn := p.socketMgr.getConnection()
		if conn == nil {
			// No connection yet, wait a bit
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-func() <-chan struct{} {
				ch := make(chan struct{})
				go func() {
					// Brief sleep before retry
					for i := 0; i < 10; i++ {
						select {
						case <-ctx.Done():
							close(ch)
							return
						default:
						}
					}
					close(ch)
				}()
				return ch
			}():
			}
			continue
		}

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				// Connection closed, wait for new connection
				continue
			}
			// Non-fatal, socket may reconnect
			continue
		}

		if n > 0 {
			if _, err := p.ptmx.Write(buf[:n]); err != nil {
				return fmt.Errorf("pty write error (from socket): %w", err)
			}
		}
	}
}

// waitForCommand waits for the command to complete and captures the exit code.
func (p *PTYTee) waitForCommand(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context cancelled, kill the process
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
		return ctx.Err()
	case err := <-errCh:
		// Command completed
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					p.exitCode = status.ExitStatus()
					p.exitCodeSet = true
				}
			}
			// Don't return error for non-zero exit codes
			if p.exitCodeSet {
				return nil
			}
			return fmt.Errorf("command error: %w", err)
		}
		p.exitCode = 0
		p.exitCodeSet = true
		return nil
	}
}

// cleanup restores terminal state and closes resources.
func (p *PTYTee) cleanup() {
	if p.termState != nil {
		p.termState.restore()
	}
	if p.ptmx != nil {
		p.ptmx.Close()
	}
	if p.socketMgr != nil {
		p.socketMgr.close()
	}
}

// SocketPath returns the Unix socket path.
func (p *PTYTee) SocketPath() string {
	return p.cfg.SocketPath
}

// ExitCode returns the exit code of the command.
// Returns 0 if the command hasn't exited yet or exited successfully.
func (p *PTYTee) ExitCode() int {
	return p.exitCode
}
