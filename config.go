package ptyee

import (
	"errors"
	"io"
	"os"
)

// OutputFormat specifies the format for data written to the socket.
type OutputFormat int

const (
	// OutputRaw writes raw bytes without any formatting.
	OutputRaw OutputFormat = iota
	// OutputJSONL writes each byte as a JSON line: {"f":"T|P","b":0-255}
	OutputJSONL
	// OutputTagged prefixes each byte with 'T' (terminal) or 'P' (program).
	OutputTagged
)

// Config holds the configuration for a PTYTee instance.
type Config struct {
	// Command is the program and arguments to run (required).
	Command []string

	// SocketPath is the Unix socket path for monitoring/injection.
	// If empty, defaults to "/tmp/ptyee.socket".
	SocketPath string

	// OutputFormat specifies how data is formatted when written to the socket.
	// Defaults to OutputRaw.
	OutputFormat OutputFormat

	// Stdin is the input source for terminal data.
	// If nil, defaults to os.Stdin.
	Stdin io.Reader

	// Stdout is the output destination for program output.
	// If nil, defaults to os.Stdout.
	Stdout io.Writer

	// Stderr is the output destination for errors.
	// If nil, defaults to os.Stderr.
	Stderr io.Writer

	// WaitForConnection, if true, waits for a client to connect to the
	// socket before starting the command. This prevents output from being
	// lost if the client connects after the program starts.
	// Defaults to true.
	WaitForConnection bool
}

// validate checks the config for required fields and sets defaults.
func (c *Config) validate() error {
	if len(c.Command) == 0 {
		return errors.New("config: command is required")
	}

	// Set defaults
	if c.SocketPath == "" {
		c.SocketPath = "/tmp/ptyee.socket"
	}

	if c.Stdin == nil {
		c.Stdin = os.Stdin
	}

	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}

	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}

	// Validate OutputFormat
	if c.OutputFormat < OutputRaw || c.OutputFormat > OutputTagged {
		return errors.New("config: invalid output format")
	}

	return nil
}
