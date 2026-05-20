package usecase

import (
	"context"

	"github.com/soi/claude-swap-widget/backend/internal/domain"
)

// SessionsReport returns the safe-to-swap report for the auto-swap state machine.
func (s *Service) SessionsReport(ctx context.Context) (*domain.SessionReport, error) {
	return s.Sessions.Report(ctx)
}

// SessionsList returns all live Claude Code sessions.
func (s *Service) SessionsList(ctx context.Context) ([]domain.ClaudeSession, error) {
	return s.Sessions.List(ctx)
}
