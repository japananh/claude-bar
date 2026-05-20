package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

func runRefreshTokens(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("refresh-tokens", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	err := svc.RefreshAllTokens(ctx)

	if *jsonOut {
		result := map[string]any{"ok": err == nil}
		if err != nil {
			result["error"] = err.Error()
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	if err != nil {
		return err
	}
	fmt.Println("All inactive account tokens refreshed.")
	return nil
}
