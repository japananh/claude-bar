package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

func runSwitch(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("switch", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: csw switch <num>")
	}
	num, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid account number: %s", fs.Arg(0))
	}
	if err := svc.SwitchAccount(ctx, num); err != nil {
		return err
	}
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"ok":              true,
			"activeAccountNumber": num,
			"hint":            "restart claude (or quit IDE plugin) to pick up the new credentials",
		})
		return nil
	}
	fmt.Printf("Switched to account %d. Restart `claude` to use the new credentials.\n", num)
	return nil
}

func runActive(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("active", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	_ = fs.Parse(args)

	res, err := svc.ListAccounts(ctx)
	if err != nil {
		return err
	}
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"activeAccountNumber": res.ActiveAccountNumber,
		})
		return nil
	}
	fmt.Println(res.ActiveAccountNumber)
	return nil
}
