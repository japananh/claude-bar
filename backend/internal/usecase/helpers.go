package usecase

import "github.com/soi/claude-swap-widget/backend/internal/domain"

// emptyConfig is a sentinel used when ~/.claude.json doesn't exist yet.
var emptyConfig = domain.ClaudeConfig{Raw: map[string]any{}}

func newOAuthAccount(email, orgName, orgUUID string) *domain.OAuthAccount {
	return &domain.OAuthAccount{
		EmailAddress:     email,
		OrganizationName: orgName,
		OrganizationUUID: orgUUID,
	}
}
