package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/soi/claude-swap-widget/backend/internal/adapter/cloudsync"
)

// CloudPush reads all accounts + backup credentials, encrypts them with the
// given passphrase, and writes the bundle to iCloud Drive.
func (s *Service) CloudPush(ctx context.Context, passphrase string) error {
	if passphrase == "" {
		return fmt.Errorf("passphrase must not be empty")
	}

	reg, err := s.Registry.Load(ctx)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	bundle := &cloudsync.CloudBundle{
		Version:  1,
		PushedAt: time.Now().UTC(),
	}

	for _, acc := range reg.Accounts {
		var blob string
		if acc.Number == reg.ActiveAccountNumber {
			// Active account: read from the live keychain entry.
			live, err := s.Live.Read(ctx)
			if err != nil || live == "" {
				continue
			}
			blob = string(live)
		} else {
			bak, err := s.Backup.Read(ctx, acc.Number, acc.Email)
			if err != nil || bak == "" {
				continue
			}
			blob = string(bak)
		}
		bundle.Accounts = append(bundle.Accounts, cloudsync.BundleAccount{
			Number:           acc.Number,
			Email:            acc.Email,
			Nickname:         acc.Nickname,
			OrganizationName: acc.OrganizationName,
			OrganizationUUID: acc.OrganizationUUID,
			CredentialBlob:   blob,
			UpdatedAt:        acc.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	if len(bundle.Accounts) == 0 {
		return fmt.Errorf("no accounts with credentials to push")
	}

	encrypted, err := cloudsync.Encrypt(bundle, passphrase)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	dest := cloudsync.BundlePath()
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return fmt.Errorf("create iCloud dir: %w", err)
	}
	if err := os.WriteFile(dest, encrypted, 0o600); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}
	return nil
}
