// Package ptyee is a tool and library for teeing a PTY.
//
// It works by creating two PTY pairs:
//
//	Terminal <--> PTY A <--> ptyee <--> PTY B <--> PROGRAM
//	                           ^
//	                           |
//	                           ⌄
//	                    Socket S (monitor/inject)
//
// The package can be used as a library to build custom PTY monitoring tools,
// or via the ptyee command-line tool.
package ptyee
