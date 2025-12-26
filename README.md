# ptyee

`ptyee` is a tool and a library for `tee`ing a `pty`.

It works like this:

```
Terminal <--> PTY A <--> ptyee <--> PTY B <--> PROGRAM
                           ^
                           |
                           ⌄
                    Socket S (monitor/inject)
```

PROGRAM is a CLI program (it could be Claude Code or something else.
Input from the terminal is sent to the program (via PTY B) and the
socket S.  Output from the program is sent to the terminal (via PTY A)
and the socket S. Input received from the socket S is sent to the
program.

## Usage

```shell
# Run COMMAND with argument ARGS in pytee
ptyee [OPTIONS] -- COMMAND [ARGS]
```

whjich runs `COMMAND` with argument `ARGS`. 

`OPTIONS` are:
* `--socket` the path to the unix socket defaults to `/tmp/ptyee.socket`
* `--jsonl-output` data written by `ptyee` to the socket is wrapped in
  a JSONL message, see JSONL below.
* `--tag-output` data written by `ptyee` to the socket is written in
  "Tag Format" (see below).
  
  
## JSONL Format

The format written to the socket when the `--jsonl-output` flag is set is:

```json
{ "f": SOURCE, "b": BYTE }
```

where:

* `SOURCE` is `T` for the terminal or `P` for the program.
* `BYTE` is a byte encoded as an integer between 0 and 255 inclusive.
  
## Tag Format

A binary format, bytes are written with a preceeding tag byte
indicating what wrote them.  E.g., the byte `A`
is written as:

```
XA
```

Where `X` is the tag: `T` for the terminal or `P` for the program.

So the string "Hello" written by the program will be:

```
PHPePlPlPo
```

The string "ls /usr" written by the terminal will be 

```
T/TuTsTr
```

Binary bytes (outside the normal ASCII) range will be written as is.

## Inspiration

This program was inspired by a conversation I had with Claude.  See
`CLAUDE_CONVERSATION.md`.




