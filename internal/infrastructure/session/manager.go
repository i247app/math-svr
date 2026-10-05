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

// InitSession stores a new, empty session under sessionKey. It returns false,
// and stores nothing, when a session already exists under that key.
func (m *SessionManager) InitSession(sessionKey string) (*AppSession, bool) {
	sess, loaded := m.SessionContainer.LoadOrStore(sessionKey, NewSession())
	if loaded {
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

// TakeSession deletes the session stored under sessionKey and returns it,
// or reports false when there is none (nothing is deleted then). The key is
// the session token exactly as the client holds it — the gex providers
// store sessions under it.
func (m *SessionManager) TakeSession(sessionKey string) (*AppSession, bool) {
	sess, ok := m.Session(sessionKey)
	if !ok {
		return nil, false
	}
	m.DeleteSession(sessionKey)
	return sess, true
}

// SetSecure sets is_secure on the session stored under sessionKey and
// returns it, or reports false when there is none. It marks the store dirty
// itself: GexSessionMiddleware only notices a change to the REQUEST's own
// session, and this one belongs to someone else.
func (m *SessionManager) SetSecure(sessionKey string, secure bool) (*AppSession, bool) {
	sess, ok := m.Session(sessionKey)
	if !ok {
		return nil, false
	}
	sess.Put("is_secure", secure)
	m.MarkDirty()
	return sess, true
}

// The loops below take each session's key from the snapshot map rather than
// from the session itself: a session does not need to know its own key.

func (m *SessionManager) DeleteAll() {
	for key := range *m.Sessions() {
		m.DeleteSession(key)
	}
}

func (m *SessionManager) MarkExpiredSessions() {
	for key, sess := range *m.Sessions() {
		if sess.IsExpired() {
			log.Printf("MarkExpiredSessions: marked session %s for deletion", ShortKey(key))
			sess.MarkForDeletion()
		}
	}
}

func (m *SessionManager) DeleteUnSecureSessions() {
	for key, sess := range *m.Sessions() {
		if !sess.IsSecure() {
			data, err := json.Marshal(sess.ToMap())
			if err != nil {
				log.Printf("DeleteUnSecureSessions: failed to marshal session %s: %v", ShortKey(key), err)
			} else {
				log.Printf("DeleteUnSecureSessions: deleting session %s | %s", ShortKey(key), data)
			}
			m.DeleteSession(key)
		}
	}
}

func (m *SessionManager) DeleteUserSessions(uid int64) {
	for _, sess := range *m.Sessions() {
		id, ok := sess.UID()
		if ok && id == uid {
			log.Printf("DeleteUserSessions: deleting session for uid is %d", id)
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
