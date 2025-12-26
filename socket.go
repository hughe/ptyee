package ptyee

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
)

// socketManager manages a Unix socket with single-client semantics.
type socketManager struct {
	listener  net.Listener
	conn      net.Conn
	connMu    sync.RWMutex
	outWriter OutputWriter
	format    OutputFormat
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

	return &socketManager{
		listener: listener,
		format:   format,
	}, nil
}

// acceptConnection accepts a single client connection.
// Additional connection attempts will be rejected with an error message.
func (sm *socketManager) acceptConnection(ctx context.Context) error {
	for {
		conn, err := sm.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return fmt.Errorf("accept error: %w", err)
			}
		}

		sm.connMu.Lock()
		if sm.conn != nil {
			// Already have a connection, reject this one
			sm.connMu.Unlock()
			conn.Write([]byte("ERROR: only one client allowed\n"))
			conn.Close()
			continue
		}

		// Accept this connection
		sm.conn = conn
		sm.outWriter = newOutputWriter(sm.format, conn)
		sm.connMu.Unlock()

		// Wait for context cancellation or connection close
		<-ctx.Done()
		return ctx.Err()
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
func (sm *socketManager) getConnection() net.Conn {
	sm.connMu.RLock()
	defer sm.connMu.RUnlock()
	return sm.conn
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
