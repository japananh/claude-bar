package keychain

import (
	"context"
	"errors"
	"os"

	"github.com/soi/claude-swap-widget/backend/internal/domain"
)

// claudeCodeService is the Keychain service name Claude Code itself uses.
const claudeCodeService = "Claude Code-credentials"

// LiveCredentialStore reads/writes the Keychain entry Claude Code reads on startup.
type LiveCredentialStore struct {
	kc *Keychain
}

// NewLiveCredentialStore binds to the current $USER on macOS Keychain.
func NewLiveCredentialStore() *LiveCredentialStore {
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	return &LiveCredentialStore{kc: New(claudeCodeService, user)}
}

// Read returns the live credential blob, or "" if not logged in.
func (s *LiveCredentialStore) Read(ctx context.Context) (domain.CredentialBlob, error) {
	out, err := s.kc.Read(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return domain.CredentialBlob(out), nil
}

// Write upserts the live credential blob.
func (s *LiveCredentialStore) Write(ctx context.Context, blob domain.CredentialBlob) error {
	return s.kc.Write(ctx, string(blob))
}

// Delete removes the live credential entry (equivalent of /logout).
func (s *LiveCredentialStore) Delete(ctx context.Context) error {
	return s.kc.Delete(ctx)
}
