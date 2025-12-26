# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ptyee` is both a CLI tool and a Go library for "teeing" a PTY (pseudo-terminal). It intercepts and monitors bidirectional communication between a terminal and a program, allowing external processes to observe and inject data via a Unix socket.

### Architecture

The core architecture creates a double-PTY proxy:

```
Terminal <--> PTY A <--> ptyee <--> PTY B <--> PROGRAM
                           ^
                           |
                           ⌄
                    Socket S (monitor/inject)
```

- **PTY A**: Connects the terminal to ptyee
- **PTY B**: Connects ptyee to the target program
- **Socket S**: Unix socket for external monitoring/injection

Data flows:
- Terminal → PTY A → ptyee → PTY B → Program → Socket S
- Socket S → ptyee → PTY B → Program
- Program → PTY B → ptyee → PTY A → Terminal → Socket S

### Code Structure

- `ptyee.go`: Library implementation (currently empty placeholder)
- `doc.go`: Package documentation
- `cmd/ptyee/main.go`: CLI entry point (not yet implemented)

This is a dual-purpose project: both an importable Go library (`github.com/hugh/ptyee`) and a standalone CLI tool.

## Common Commands

### Building
```bash
make build          # Build CLI to build/ptyee
make install        # Install CLI to $GOPATH/bin
```

### Testing
```bash
make test           # Run all tests with verbose output
go test ./...       # Standard Go test runner
```

### Code Quality
```bash
make fmt            # Format all Go code
make vet            # Run go vet
make lint           # Run golangci-lint (requires golangci-lint installed)
```

### Running
```bash
make run            # Build and run the CLI
# Once implemented:
# ptyee [OPTIONS] -- COMMAND [ARGS]
```

## CLI Design

The CLI will support these options:

- `--socket PATH`: Unix socket path (default: `/tmp/ptyee.socket`)
- `--jsonl-output`: Output JSONL format to socket
- `--tag-output`: Output tag format to socket

### Output Formats

**JSONL Format** (`--jsonl-output`):
```json
{ "f": "T|P", "b": 0-255 }
```
- `f`: Source (`T` for terminal, `P` for program)
- `b`: Byte value (0-255)

**Tag Format** (`--tag-output`):
Binary format where each byte is prefixed with a tag (`T` or `P`).
Example: "Hello" from program → `PHPePlPlPo`

## Implementation Status

The project structure is initialized but core functionality is not yet
implemented. The main.go currently exits with "not yet implemented"
error.

# IMPORTANT: Task Tracking

This project uses **bd** (beads) for issue tracking.  **Do not** use
 markdown files or the todo list.  Run `bd quickstart` to learn how.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. 

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Hand off** - Provide context for next session

**DO NOT PUSH** wait for the human to do a code review.  The human will push.


