package ptyee

import (
	"io"
)

// OutputWriter formats and writes data to the socket based on the configured format.
type OutputWriter interface {
	// WriteFromTerminal writes a byte that originated from the terminal.
	WriteFromTerminal(b byte) error
	// WriteFromProgram writes a byte that originated from the program.
	WriteFromProgram(b byte) error
}

// rawOutputWriter writes bytes without any formatting.
type rawOutputWriter struct {
	w io.Writer
}

func (r *rawOutputWriter) WriteFromTerminal(b byte) error {
	_, err := r.w.Write([]byte{b})
	return err
}

func (r *rawOutputWriter) WriteFromProgram(b byte) error {
	_, err := r.w.Write([]byte{b})
	return err
}

// taggedOutputWriter prefixes each byte with 'T' (terminal) or 'P' (program).
type taggedOutputWriter struct {
	w io.Writer
}

func (t *taggedOutputWriter) WriteFromTerminal(b byte) error {
	_, err := t.w.Write([]byte{'T', b})
	return err
}

func (t *taggedOutputWriter) WriteFromProgram(b byte) error {
	_, err := t.w.Write([]byte{'P', b})
	return err
}

// newOutputWriter creates an OutputWriter based on the specified format.
// Returns nil if the format is not yet implemented (e.g., OutputJSONL).
func newOutputWriter(format OutputFormat, w io.Writer) OutputWriter {
	switch format {
	case OutputRaw:
		return &rawOutputWriter{w: w}
	case OutputTagged:
		return &taggedOutputWriter{w: w}
	case OutputJSONL:
		// TODO: Implement JSONL output format (deferred)
		return nil
	default:
		return nil
	}
}
