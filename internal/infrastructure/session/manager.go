package session

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/i247app/gex/session"
)

type SessionRequestContextKey string

const (
	SessionContextKey = SessionRequestContextKey("math_session")
)

func Dump(sessionManager *SessionManager) *map[string]map[string]any {
	result := make(map[string]map[string]any)

	sessions := sessionManager.Sessions()
	for k, v := range *sessions {
		result[k] = v.ToMap()
		// log.Printf(">> Dump: k=%s\tv=%v\n", k, result[k])
	}

	return &result
}

func GetRequestSession(r *http.Request) *AppSession {
	val := r.Context().Value(SessionContextKey)
	if val == nil {
		return nil
	}

	storer, ok := val.(session.SessionStorer)
	if !ok {
		return nil
	}

	sess, ok := storer.(*AppSession)
	if !ok {
		return nil
	}

	return sess
}

// type SessionManager interface {
// 	Container() *session.Container
// 	Session(sessionKey string) (*AppSession, bool)
// 	Sessions() *map[string]*AppSession
// 	InitSession(sessionKey string) (*AppSession, bool)
// 	DeleteSession(sessionKey string)
// }

type SessionManager struct {
	SessionContainer session.Container

	// persister is nil unless write-through persistence is enabled
	// (EnablePersistence); every method that uses it is nil-safe.
	persister *persister
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		SessionContainer: *session.NewContainer(),
	}
}

func (m *SessionManager) Container() *session.Container {
	return &m.SessionContainer
}

func (m *SessionManager) Session(sessionKey string) (*AppSession, bool) {
	sess, ok := m.SessionContainer.Session(sessionKey)
	if !ok {
		return nil, false
	}

	monexSess, ok := sess.(*AppSession)
	if !ok {
		return nil, false
	}

	return monexSess, true
}

// Sessions returns a point-in-time copy of the container, taken under its
// lock, so callers may iterate it while requests keep running.
func (m *SessionManager) Sessions() *map[string]*AppSession {
	sessions := m.SessionContainer.Snapshot()
	result := make(map[string]*AppSession, len(sessions))
	for k, v := range sessions {
		monexSess, ok := v.(*AppSession)
		if !ok {
			continue
		}
		result[k] = monexSess
	}

	return &result
}

func (m *SessionManager) InitSession(sessionKey string) (*AppSession, bool) {
	sess, ok := m.SessionContainer.InitSession(sessionKey, NewSession())
	if !ok {
		return nil, false
	}

	monexSess, ok := sess.(*AppSession)
	if !ok {
		return monexSess, false
	}

	return monexSess, true
}

func (m *SessionManager) DeleteSession(sessionKey string) {
	m.SessionContainer.DeleteSession(sessionKey)
	m.MarkDirty()
}

func (m *SessionManager) DeleteAll() {
	for _, sess := range *m.Sessions() {
		key, ok := sess.Get("key")
		if !ok {
			continue
		}
		m.DeleteSession(key.(string))
	}
}

func (m *SessionManager) MarkExpiredSessions() {
	for _, sess := range *m.Sessions() {
		key, ok := sess.Get("key")
		if !ok {
			continue
		}

		if sess.IsExpired() {
			log.Printf("MarkExpiredSessions: marked session %s for deletion", ShortKey(key.(string)))
			sess.MarkForDeletion()
		}
	}
}

func (m *SessionManager) DeleteUnSecureSessions() {
	for _, sess := range *m.Sessions() {
		key, ok := sess.Get("key")
		if !ok {
			continue
		}

		if !sess.IsSecure() {
			data, err := json.Marshal(sess.ToMap())
			if err != nil {
				log.Printf("DeleteUnSecureSessions: failed to marshal session %s: %v", ShortKey(key.(string)), err)
			} else {
				log.Printf("DeleteUnSecureSessions: deleting session %s | %s", ShortKey(key.(string)), data)
			}
			m.DeleteSession(key.(string))
		}
	}
}

func (m *SessionManager) DeleteUserSessions(uid int64) {
	for _, sess := range *m.Sessions() {
		id, ok := sess.UID()
		if ok && id == uid {
			log.Printf("DeleteUserSessions: deleting session for userID is %d", id)
			// sess.MarkForDeletion()
			sess.MarkNotSecure()
		}
	}
	m.MarkDirty()
}

// ShortKey shortens a session key for logging. Keys are signed session
// tokens, i.e. live credentials, so a full key must never reach a log line.
func ShortKey(key string) string {
	if len(key) > 19 {
		return key[:8] + "..." + key[len(key)-8:]
	}
	return key
}
