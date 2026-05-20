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

func runAdd(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	nickname := fs.String("nickname", "", "optional display name for the account")
	_ = fs.Parse(args)

	res, err := svc.AddAccount(ctx, *nickname)
	if err != nil {
		return err
	}
	if *asJSON {
		return json.NewEncoder(os.Stdout).Encode(res)
	}
	if res.WasDuplicate {
		fmt.Printf("⚠ Account %s already existed as Account-%d. Backup credentials refreshed.\n",
			res.Account.Email, res.DuplicateOfNum)
	} else {
		fmt.Printf("Added Account-%d (%s).\n", res.Account.Number, res.Account.DisplayName())
	}
	return nil
}

func runRename(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("rename", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: csw rename <num> [<nickname>]   (empty nickname clears)")
	}
	num, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid account number: %s", fs.Arg(0))
	}
	nickname := ""
	if fs.NArg() >= 2 {
		nickname = fs.Arg(1)
	}
	if err := svc.RenameAccount(ctx, num, nickname); err != nil {
		return err
	}
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"ok":       true,
			"number":   num,
			"nickname": nickname,
		})
		return nil
	}
	if nickname == "" {
		fmt.Printf("Cleared nickname for Account-%d.\n", num)
	} else {
		fmt.Printf("Renamed Account-%d to %q.\n", num, nickname)
	}
	return nil
}

func runRemove(ctx context.Context, svc *usecase.Service, args []string) error {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: csw remove <num>")
	}
	num, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid account number: %s", fs.Arg(0))
	}
	if err := svc.RemoveAccount(ctx, num); err != nil {
		return err
	}
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "number": num})
		return nil
	}
	fmt.Printf("Removed Account-%d.\n", num)
	return nil
}
