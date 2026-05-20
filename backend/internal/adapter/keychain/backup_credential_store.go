package keychain

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/soi/claude-swap-widget/backend/internal/domain"
)

// backupServiceFormat keeps each inactive account in its own Keychain entry,
// scoped by widget so it never collides with Claude Code's live entry.
const backupServiceFormat = "csw-backup:%d:%s"

// BackupCredentialStore holds credentials for inactive accounts.
type BackupCredentialStore struct {
	user string
}

// NewBackupCredentialStore binds to the current $USER.
func NewBackupCredentialStore() *BackupCredentialStore {
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	return &BackupCredentialStore{user: user}
}

func (s *BackupCredentialStore) kc(accountNum int, email string) *Keychain {
	return New(fmt.Sprintf(backupServiceFormat, accountNum, email), s.user)
}

// Read returns the backed-up credential for an inactive account.
func (s *BackupCredentialStore) Read(ctx context.Context, accountNum int, email string) (domain.CredentialBlob, error) {
	out, err := s.kc(accountNum, email).Read(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return domain.CredentialBlob(out), nil
}

// Write saves a backup credential for an inactive account.
func (s *BackupCredentialStore) Write(ctx context.Context, accountNum int, email string, blob domain.CredentialBlob) error {
	return s.kc(accountNum, email).Write(ctx, string(blob))
}

// Delete removes a backup credential.
func (s *BackupCredentialStore) Delete(ctx context.Context, accountNum int, email string) error {
	return s.kc(accountNum, email).Delete(ctx)
}
