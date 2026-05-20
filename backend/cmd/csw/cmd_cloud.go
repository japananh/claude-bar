package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/soi/claude-swap-widget/backend/internal/usecase"
)

// runCloud dispatches cloud sub-commands: status | push | pull | forget
func runCloud(ctx context.Context, svc *usecase.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: csw cloud <status|push|pull|forget> [--json]")
	}

	jsonOut := len(args) > 1 && args[len(args)-1] == "--json"
	sub := args[0]

	switch sub {
	case "status":
		res, err := svc.CloudStatus(ctx)
		if err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(res)
		}
		if !res.Exists {
			fmt.Printf("No bundle found at:\n  %s\n", res.Path)
			return nil
		}
		fmt.Printf("Bundle: %s\nLast pushed: %s  (%d KB)\n",
			res.Path, res.PushedAt.Format("2006-01-02 15:04:05 UTC"), res.SizeKB)
		return nil

	case "push":
		pass, err := readPassphrase("Passphrase: ")
		if err != nil {
			return err
		}
		if err := svc.CloudPush(ctx, pass); err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true})
		}
		fmt.Println("Bundle pushed to iCloud Drive.")
		return nil

	case "pull":
		pass, err := readPassphrase("Passphrase: ")
		if err != nil {
			return err
		}
		if err := svc.CloudPull(ctx, pass); err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true})
		}
		fmt.Println("Accounts restored from iCloud Drive.")
		return nil

	case "forget":
		if err := svc.CloudForget(ctx); err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true})
		}
		fmt.Println("Bundle removed from iCloud Drive.")
		return nil

	default:
		return fmt.Errorf("unknown cloud sub-command: %s", sub)
	}
}

// readPassphrase reads a line from stdin (used when called interactively or
// with a pipe from Swift: passphrase written to the process stdin pipe).
func readPassphrase(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("no passphrase provided")
	}
	pass := strings.TrimRight(scanner.Text(), "\r\n")
	if pass == "" {
		return "", fmt.Errorf("passphrase must not be empty")
	}
	return pass, nil
}
