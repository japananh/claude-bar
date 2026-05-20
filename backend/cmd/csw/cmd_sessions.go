package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

func runSessions(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("sessions", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	verbose := fs.Bool("v", false, "list each session")
	_ = fs.Parse(args)

	rep, err := svc.SessionsReport(ctx)
	if err != nil {
		return err
	}
	if *asJSON {
		out := map[string]any{"report": rep}
		if *verbose {
			list, _ := svc.SessionsList(ctx)
			out["sessions"] = list
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}
	fmt.Printf("Live sessions:    %d\n", rep.Total)
	fmt.Printf("Busy or waiting:  %d\n", rep.BusyOrWaiting)
	fmt.Printf("Interactive:      %d\n", rep.InteractiveOnly)
	if rep.SafeToSwap {
		fmt.Println("✓ Safe to swap")
	} else {
		fmt.Println("✗ Wait — claude is running")
	}
	if *verbose {
		list, _ := svc.SessionsList(ctx)
		for _, s := range list {
			fmt.Printf("  pid=%-6d kind=%-12s status=%-7s cwd=%s\n", s.PID, s.Kind, s.Status, s.CWD)
		}
	}
	return nil
}
