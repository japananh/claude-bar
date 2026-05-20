package usecase

import (
	"context"
	"fmt"
	"strings"
)

// RefreshAllTokens proactively refreshes the OAuth credentials for every
// inactive account. Called once per day by the widget on startup so that
// backup tokens never go stale between swaps.
//
// The active account is intentionally skipped — Claude Code owns its token
// refresh cycle while it is running.
func (s *Service) RefreshAllTokens(ctx context.Context) error {
	reg, err := s.Registry.Load(ctx)
	if err != nil {
		return err
	}

	var failures []string
	for num, acc := range reg.Accounts {
		if num == reg.ActiveAccountNumber {
			continue
		}
		blob, err := s.Backup.Read(ctx, acc.Number, acc.Email)
		if err != nil || blob == "" {
			continue
		}
		payload, err := blob.Extract()
		if err != nil || payload.RefreshToken == "" {
			continue
		}
		fresh, err := s.Refresh.Refresh(ctx, payload.RefreshToken)
		if err != nil {
			failures = append(failures, fmt.Sprintf("account %d (%s): %v", num, acc.Email, err))
			continue
		}
		if fresh == nil {
			continue
		}
		refreshed, err := blob.WithRefreshed(fresh)
		if err != nil || refreshed == "" {
			continue
		}
		_ = s.Backup.Write(ctx, acc.Number, acc.Email, refreshed)
	}
	if len(failures) > 0 {
		return fmt.Errorf("partial refresh failures: %s", strings.Join(failures, "; "))
	}
	return nil
}
