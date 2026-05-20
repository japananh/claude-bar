package usecase

import (
	"context"
	"fmt"
	"strings"
)

// RenameAccount updates the nickname for an account.
// Empty nickname clears it (display falls back to email).
func (s *Service) RenameAccount(ctx context.Context, num int, nickname string) error {
	if err := s.Lock.Acquire(ctx); err != nil {
		return err
	}
	defer s.Lock.Release()

	reg, err := s.Registry.Load(ctx)
	if err != nil {
		return err
	}
	acc, ok := reg.Accounts[num]
	if !ok {
		return fmt.Errorf("account %d not found", num)
	}
	acc.Nickname = strings.TrimSpace(nickname)
	return s.Registry.Save(ctx, reg)
}
