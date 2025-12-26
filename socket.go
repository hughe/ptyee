package ptyee

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// socketManager manages a Unix socket with single-client semantics.
type socketManager struct {
	listener      *net.UnixListener
	conn          *net.UnixConn
	connMu        sync.RWMutex
	outWriter     OutputWriter
	format        OutputFormat
	connReady     chan struct{} // Signals when first connection is established
	connReadyOnce sync.Once     // Ensures connReady is closed only once
}

// newSocketManager creates a socket manager and starts listening on the specified path.
func newSocketManager(socketPath string, format OutputFormat) (*socketManager, error) {
	// Remove existing socket file if present
	if err := os.RemoveAll(socketPath); err != nil {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %w", err)
	}

	unixListener, ok := listener.(*net.UnixListener)
	if !ok {
		return nil, fmt.Errorf("expected UnixListener, got %T", listener)
	}

	// Automatically unlink socket file on close
	unixListener.SetUnlinkOnClose(true)

	return &socketManager{
		listener:  unixListener,
		format:    format,
		connReady: make(chan struct{}),
	}, nil
}

// acceptConnection accepts a single client connection and continues to listen.
// Additional connection attempts will be silently closed without error messages.
func (sm *socketManager) acceptConnection(ctx context.Context) error {
	// Set accept deadline for periodic context checking

	for {
		// Use a moderate deadline to check context periodically without too much overhead
		sm.listener.SetDeadline(time.Now().Add(100 * time.Millisecond))

		unixConn, err := sm.listener.AcceptUnix()
		if err != nil {
			// Check if it's a timeout error (expected)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check context after timeout
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					continue // Timeout is expected, try again
				}
			}
			// Non-timeout error
			return fmt.Errorf("accept error: %w", err)
		}

		sm.connMu.Lock()
		if sm.conn != nil {
			// Already have a connection, close this one without writing error
			sm.connMu.Unlock()
			unixConn.Close()
			continue
		}

		// Accept this connection
		sm.conn = unixConn
		sm.outWriter = newOutputWriter(sm.format, unixConn)

		// Signal that first connection is ready
		sm.connReadyOnce.Do(func() {
			close(sm.connReady)
		})

		sm.connMu.Unlock()

		// Continue accepting (and rejecting) additional connections
	}
}

// writeFromTerminal writes a byte that originated from the terminal to the connected client.
func (sm *socketManager) writeFromTerminal(b byte) error {
	sm.connMu.RLock()
	defer sm.connMu.RUnlock()

	if sm.outWriter == nil {
		return nil // No client connected, silently ignore
	}

	return sm.outWriter.WriteFromTerminal(b)
}

// writeFromProgram writes a byte that originated from the program to the connected client.
func (sm *socketManager) writeFromProgram(b byte) error {
	sm.connMu.RLock()
	defer sm.connMu.RUnlock()

	if sm.outWriter == nil {
		return nil // No client connected, silently ignore
	}

	return sm.outWriter.WriteFromProgram(b)
}

// getConnection returns the current client connection for reading.
// Returns nil if no client is connected.
func (sm *socketManager) getConnection() *net.UnixConn {
	sm.connMu.RLock()
	defer sm.connMu.RUnlock()
	return sm.conn
}

// waitForFirstConnection blocks until the first client connects to the socket.
// Returns an error if the context is cancelled before a connection is established.
func (sm *socketManager) waitForFirstConnection(ctx context.Context) error {
	select {
	case <-sm.connReady:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// close closes the socket manager and cleans up resources.
func (sm *socketManager) close() error {
	sm.connMu.Lock()
	if sm.conn != nil {
		sm.conn.Close()
		sm.conn = nil
		sm.outWriter = nil
	}
	sm.connMu.Unlock()

	if sm.listener != nil {
		return sm.listener.Close()
	}
	return nil
}
