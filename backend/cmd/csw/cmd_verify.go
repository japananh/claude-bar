package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/soi/claude-swap-widget/backend/internal/domain"
	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

func runVerify(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	_ = fs.Parse(args)

	report, err := svc.VerifyAccounts(ctx)
	if err != nil {
		return err
	}
	if *asJSON {
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	printVerifyReport(report)
	return nil
}

func printVerifyReport(r *domain.VerificationReport) {
	if r.Total == 0 {
		fmt.Println("No accounts to verify.")
		return
	}
	fmt.Printf("Verifying %d account(s)…\n\n", r.Total)
	for _, res := range r.Results {
		active := ""
		if res.IsActive {
			active = " (ACTIVE)"
		}
		fmt.Printf("  Account-%d  %-24s  %s%s\n",
			res.AccountNum, res.DisplayName, res.Email, active)
		for _, c := range res.Checks {
			fmt.Printf("    %s %s%s\n", checkGlyph(c), c.Name, detailSuffix(c))
		}
		if res.SwapReady {
			fmt.Println("    → Swap ready ✓")
		} else {
			fmt.Println("    → Cannot swap — re-add this account")
		}
		fmt.Println()
	}
	fmt.Printf("Summary: %d of %d account(s) swap-ready\n", r.Ready, r.Total)
}

func checkGlyph(c domain.CheckResult) string {
	switch {
	case c.Skipped: return "-"
	case c.Passed:  return "✓"
	default:        return "✗"
	}
}

func detailSuffix(c domain.CheckResult) string {
	if c.Detail == "" {
		return ""
	}
	return ": " + c.Detail
}
