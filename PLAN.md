# Refactoring Plan: ptyee Library

## Overview
Refactor ptyee to address issues ptyee-7q8, ptyee-56x, and ptyee-opa:
1. Move main() from library to cmd/ptyee/main.go
2. Remove MultiWriter (support single client only)
3. Add context.Context for cancellation
4. Implement CLI with --socket and --tag-output flags (defer --jsonl-output)

## Library API Design

### Core Types
```go
// ptyee.go
type PTYTee struct {
    cmd         *exec.Cmd
    ptmx        *os.File
    socketMgr   *socketManager
    exitCode    int
    // ... other internal state
}

type Config struct {
    Command      []string
    SocketPath   string
    OutputFormat OutputFormat
    Stdin        io.Reader    // default: os.Stdin
    Stdout       io.Writer    // default: os.Stdout
    Stderr       io.Writer    // default: os.Stderr
}

type OutputFormat int
const (
    OutputRaw OutputFormat = iota
    OutputJSONL
    OutputTagged
)

// Public API
func NewPTYTee(cfg Config) (*PTYTee, error)
func (p *PTYTee) Run(ctx context.Context) error
func (p *PTYTee) SocketPath() string
func (p *PTYTee) ExitCode() int
```

## Implementation Steps

### Phase 1: Create New Files

#### 1. Create `config.go`
- Define `Config` struct
- Define `OutputFormat` constants (OutputRaw, OutputJSONL, OutputTagged)
- Add config validation function

#### 2. Create `output.go`
- Define `OutputWriter` interface with methods:
  - `WriteFromTerminal(b byte) error`
  - `WriteFromProgram(b byte) error`
- Implement two writers (defer JSONL for later):
  - `rawOutputWriter` - no formatting
  - `taggedOutputWriter` - prefix each byte with 'T' or 'P'
- Add factory: `newOutputWriter(format OutputFormat, conn net.Conn) OutputWriter`

#### 3. Create `socket.go`
- Define `socketManager` struct:
  ```go
  type socketManager struct {
      listener   net.Listener
      conn       net.Conn          // single connection
      connMu     sync.RWMutex
      outWriter  OutputWriter
  }
  ```
- Implement single-client accept logic (reject additional connections)
- Thread-safe write methods using OutputWriter

#### 4. Create `terminal.go`
- Extract terminal setup from current main():
  - Raw mode setup/restore
  - Window size handling
  - SIGWINCH signal handling
- Make functions reusable by PTYTee

### Phase 2: Refactor Core Library

#### 5. Refactor `ptyee.go`
**Remove:**
- MultiWriter struct (lines 19-62)
- main() function (lines 64-161)

**Add:**
- PTYTee struct with internal state
- `NewPTYTee(cfg Config) (*PTYTee, error)` - constructor that:
  - Validates config
  - Creates socket listener
  - Returns PTYTee instance
- `Run(ctx context.Context) error` - main execution:
  - Start command with PTY
  - Setup terminal (raw mode, window size)
  - Use errgroup.WithContext for goroutine coordination:
    - Socket accept loop
    - PTY output → stdout + socket
    - stdin → PTY
    - socket → PTY
  - Handle context cancellation gracefully
  - Wait for command completion
  - Cleanup (restore terminal, close socket)
- `SocketPath() string` - getter
- `ExitCode() int` - getter

**Key Changes:**
- Use errgroup for goroutine management
- Proper context propagation
- Return errors instead of log.Fatal
- Default socket path: `/tmp/ptyee.socket` (not `/tmp/claude.sock`)

### Phase 3: Implement CLI

#### 6. Rewrite `cmd/ptyee/main.go`
**Replace entire file with:**
```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "ptyee: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    // Parse flags (stdlib flag package)
    socketPath := flag.String("socket", "/tmp/ptyee.socket", "...")
    tagOutput := flag.Bool("tag-output", false, "...")
    flag.Parse()

    // Parse command args (after --)
    // Determine OutputFormat (Raw or Tagged)

    // Build Config
    cfg := ptyee.Config{...}

    // Create PTYTee
    pt, err := ptyee.NewPTYTee(cfg)

    // Setup context with signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    // Run
    if err := pt.Run(ctx); err != nil {
        return err
    }

    // Exit with child's exit code
    os.Exit(pt.ExitCode())
}
```

## File Changes Summary

**New Files:**
- `/Users/hugh/src/ptyee/config.go` - Config, OutputFormat
- `/Users/hugh/src/ptyee/output.go` - OutputWriter implementations
- `/Users/hugh/src/ptyee/socket.go` - socketManager (single client)
- `/Users/hugh/src/ptyee/terminal.go` - Terminal utilities

**Modified Files:**
- `/Users/hugh/src/ptyee/ptyee.go` - Remove MultiWriter/main, add PTYTee API
- `/Users/hugh/src/ptyee/cmd/ptyee/main.go` - Complete rewrite with flags

**Dependencies to Add:**
- `golang.org/x/sync/errgroup` - for goroutine coordination

## Key Design Decisions

1. **Single Client Only**: socketManager accepts one connection, rejects others
2. **Context-based Cancellation**: Use errgroup.WithContext for clean shutdown
3. **Byte-by-byte Output Formatting**: Required for Tag format per spec
4. **Deferred JSONL Format**: Focus on core refactoring first, add JSONL later
5. **No Breaking Changes**: This is the first proper release
6. **Library Returns Errors**: Don't log.Fatal in library code, let CLI handle

## Testing Approach

After implementation:
1. Build: `make build`
2. Test basic execution: `./build/ptyee -- ls -la`
3. Test socket connection: `./build/ptyee --socket /tmp/test.sock -- bash`
4. Test tag output: `./build/ptyee --tag-output -- echo hello`
5. Test context cancellation: Ctrl-C during execution
6. Test single client: Try connecting twice to socket

## Success Criteria

- [ ] No main() in ptyee.go
- [ ] MultiWriter removed
- [ ] CLI supports --socket and --tag-output flags
- [ ] Context cancellation works cleanly
- [ ] Single client enforced
- [ ] Exit code propagated correctly
- [ ] Terminal state restored on exit
- [ ] All three issues closed: ptyee-7q8, ptyee-56x, ptyee-opa

## Deferred Work

### JSONL Output Format
To be implemented in a future iteration after the core refactoring is complete.

**Requirements:**
- Add `--jsonl-output` CLI flag
- Implement `jsonlOutputWriter` in `output.go`:
  ```go
  type jsonlOutputWriter struct {
      conn net.Conn
      enc  *json.Encoder
  }

  func (j *jsonlOutputWriter) WriteFromTerminal(b byte) error {
      return j.enc.Encode(map[string]interface{}{
          "f": "T",
          "b": int(b),
      })
  }

  func (j *jsonlOutputWriter) WriteFromProgram(b byte) error {
      return j.enc.Encode(map[string]interface{}{
          "f": "P",
          "b": int(b),
      })
  }
  ```
- Update flag validation to ensure `--jsonl-output` and `--tag-output` are mutually exclusive
- Add test: `./build/ptyee --jsonl-output -- echo hello`
