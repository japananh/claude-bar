// Package registry persists the widget registry as JSON on disk with
// atomic writes and 0600 perms.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/soi/claude-swap-widget/backend/internal/adapter"
	"github.com/soi/claude-swap-widget/backend/internal/domain"
)

// FileRegistry stores the Registry as a JSON file under ~/Library/Application Support.
type FileRegistry struct {
	path string
}

// New returns a FileRegistry at the default widget data path.
func New() *FileRegistry {
	return &FileRegistry{path: adapter.RegistryFile()}
}

// NewAt returns a FileRegistry at a custom path (tests).
func NewAt(path string) *FileRegistry { return &FileRegistry{path: path} }

// Load returns the on-disk registry, or an empty one if the file doesn't exist.
func (r *FileRegistry) Load(ctx context.Context) (*domain.Registry, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.NewRegistry(), nil
		}
		return nil, err
	}
	var reg domain.Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	if reg.Accounts == nil {
		reg.Accounts = map[int]*domain.Account{}
	}
	if reg.Sequence == nil {
		reg.Sequence = []int{}
	}
	return &reg, nil
}

// Save atomically writes the registry with 0600 perms.
func (r *FileRegistry) Save(ctx context.Context, reg *domain.Registry) error {
	if err := adapter.EnsureDataDir(); err != nil {
		return err
	}
	reg.LastUpdated = time.Now().UTC()
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	tmp, err := os.CreateTemp(dir, "registry-*.tmp")
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
	return os.Rename(tmpName, r.path)
}
