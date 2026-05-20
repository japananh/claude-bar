// Package claudeconfig reads and writes ~/.claude.json, preserving the
// many unrelated fields Claude Code stores there.
package claudeconfig

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/soi/claude-swap-widget/backend/internal/adapter"
	"github.com/soi/claude-swap-widget/backend/internal/domain"
)

// ClaudeConfigStore is the on-disk adapter for ~/.claude.json.
type ClaudeConfigStore struct {
	path string
}

// New returns a store bound to ~/.claude.json.
func New() *ClaudeConfigStore {
	return &ClaudeConfigStore{path: adapter.ClaudeConfigFile()}
}

// NewAt is for tests.
func NewAt(path string) *ClaudeConfigStore { return &ClaudeConfigStore{path: path} }

// Exists reports whether the file is on disk.
func (s *ClaudeConfigStore) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

// Read returns the parsed config plus the raw map so unknown fields survive a Write.
func (s *ClaudeConfigStore) Read(ctx context.Context) (*domain.ClaudeConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &domain.ClaudeConfig{Raw: map[string]any{}}, nil
		}
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	cfg := &domain.ClaudeConfig{Raw: raw}
	if oa, ok := raw["oauthAccount"]; ok {
		b, _ := json.Marshal(oa)
		var acct domain.OAuthAccount
		if err := json.Unmarshal(b, &acct); err == nil {
			cfg.OAuthAccount = &acct
		}
	}
	return cfg, nil
}

// Write atomically replaces ~/.claude.json with OAuthAccount swapped in,
// keeping all other fields intact.
func (s *ClaudeConfigStore) Write(ctx context.Context, cfg *domain.ClaudeConfig) error {
	if cfg.Raw == nil {
		cfg.Raw = map[string]any{}
	}
	if cfg.OAuthAccount != nil {
		cfg.Raw["oauthAccount"] = cfg.OAuthAccount
	}
	data, err := json.MarshalIndent(cfg.Raw, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	tmp, err := os.CreateTemp(dir, ".claude-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}
