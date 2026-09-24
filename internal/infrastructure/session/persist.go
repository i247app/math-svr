package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/utils"
)

// persister writes the session file in the background whenever the auth
// state of a session changes, so a crash that skips OnShutdown does not lose
// who is signed in. Requests only raise a flag (MarkDirty); a single goroutine
// does the writing, so a burst of logins costs one write, not one per login.
type persister struct {
	file  string
	dirty chan struct{} // buffered(1): a pending signal already covers any later change
	stop  chan struct{}
	done  chan struct{}
}

// EnablePersistence starts the background writer for file. Call it once, at
// boot, after sessions have been reloaded; a later MarkDirty then triggers a
// snapshot write.
func (m *SessionManager) EnablePersistence(file string) {
	p := &persister{
		file:  file,
		dirty: make(chan struct{}, 1),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	m.persister = p

	go func() {
		defer close(p.done)
		for {
			select {
			case <-p.dirty:
				if err := m.SaveTo(p.file); err != nil {
					logger.From(context.Background()).Errorf("session.persist_failed err=%v", err)
				}
			case <-p.stop:
				return
			}
		}
	}()
}

// StopPersistence stops the background writer and waits for an in-flight
// write to finish. The caller writes the final snapshot itself (SaveTo).
// No-op when persistence was never enabled.
func (m *SessionManager) StopPersistence() {
	if m.persister == nil {
		return
	}
	close(m.persister.stop)
	<-m.persister.done
}

// MarkDirty asks the background writer to persist the sessions. It never
// blocks. No-op when persistence is disabled.
func (m *SessionManager) MarkDirty() {
	if m == nil || m.persister == nil {
		return
	}
	select {
	case m.persister.dirty <- struct{}{}:
	default: // a write is already pending and will include this change
	}
}

// SaveTo writes every session that belongs to a user (uid set: signed in,
// mid-login, or guest) to file. Anonymous sessions are skipped: a client that
// still holds a valid token gets an equivalent one back from the provider.
//
// The write is atomic (temp file + fsync + rename), so a crash mid-write
// leaves the previous file intact instead of a truncated one. The file holds
// live session tokens, hence 0600.
func (m *SessionManager) SaveTo(file string) error {
	data := make(map[string]any)
	for key, sess := range *m.Sessions() {
		if _, ok := sess.UID(); !ok {
			continue
		}
		data[key] = sess.ToMap()
	}

	raw, err := utils.SerializeMap(&data)
	if err != nil {
		return fmt.Errorf("session persist: serialize: %w", err)
	}

	dir := filepath.Dir(file)
	tmp, err := os.CreateTemp(dir, filepath.Base(file)+".tmp-*")
	if err != nil {
		return fmt.Errorf("session persist: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("session persist: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("session persist: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("session persist: close: %w", err)
	}
	if err := os.Rename(tmpName, file); err != nil {
		return fmt.Errorf("session persist: rename: %w", err)
	}

	// Make the rename itself durable. Best-effort: not every platform can
	// fsync a directory.
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	pending := 0
	if m.persister != nil { // nil when persistence is off (shutdown-only writes)
		pending = len(m.persister.dirty)
	}
	logger.From(context.Background()).Infof("session.persisted count=%d, length_dirty=%d", len(data), pending)
	return nil
}
