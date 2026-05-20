// Package lock provides an advisory file lock to serialise swap operations
// across processes (widget + CLI).
package lock

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/soi/claude-swap-widget/backend/internal/adapter"
)

// ErrAcquireTimeout is returned when the context expires before acquiring.
var ErrAcquireTimeout = errors.New("file lock acquire timeout")

// FileLock holds an flock on a sentinel file.
type FileLock struct {
	path string
	f    *os.File
}

// New returns a FileLock at the default widget lock path.
func New() *FileLock {
	return &FileLock{path: adapter.LockFile()}
}

// NewAt is for tests.
func NewAt(path string) *FileLock { return &FileLock{path: path} }

// Acquire blocks until the lock is held or ctx expires.
func (l *FileLock) Acquire(ctx context.Context) error {
	if err := adapter.EnsureDataDir(); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			l.f = f
			return nil
		}
		if err != syscall.EWOULDBLOCK {
			f.Close()
			return err
		}
		select {
		case <-ctx.Done():
			f.Close()
			return ErrAcquireTimeout
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Release unlocks and closes the sentinel file. Safe to call once.
func (l *FileLock) Release() error {
	if l.f == nil {
		return nil
	}
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	err := l.f.Close()
	l.f = nil
	return err
}
