// Command csw is the Claude Swap Widget backend CLI.
//
// It is invoked both standalone (by humans) and as a subprocess by the Swift
// widget. Every subcommand supports --json for machine-readable output.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	svc := usecase.NewMacOSService()
	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "list":
		err = runList(ctx, svc, args)
	case "switch":
		err = runSwitch(ctx, svc, args)
	case "add":
		err = runAdd(ctx, svc, args)
	case "rename":
		err = runRename(ctx, svc, args)
	case "remove":
		err = runRemove(ctx, svc, args)
	case "sessions":
		err = runSessions(ctx, svc, args)
	case "active":
		err = runActive(ctx, svc, args)
	case "verify":
		err = runVerify(ctx, svc, args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `csw — Claude Swap Widget backend

Commands:
  list                    List all managed accounts with usage
  active                  Print the active account number
  switch <num>            Switch active account to <num>
  add [--nickname=NAME]   Snapshot the currently-logged-in account
  rename <num> <nickname> Rename an account (empty string clears)
  remove <num>            Remove an account (must not be active)
  sessions                Report live Claude Code sessions
  verify                  Verify every account is swap-ready
  help                    Show this help

All commands accept --json for machine-readable output.`)
}
