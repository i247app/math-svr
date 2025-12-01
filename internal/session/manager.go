package session

import (
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

	mathSess, ok := sess.(*AppSession)
	if !ok {
		return nil, false
	}

	return mathSess, true
}

func (m *SessionManager) Sessions() *map[string]*AppSession {
	sessions := m.SessionContainer.Sessions()
	result := make(map[string]*AppSession)
	for k, v := range *sessions {
		mathSess, ok := v.(*AppSession)
		if !ok {
			continue
		}
		result[k] = mathSess
	}

	return &result
}

func (m *SessionManager) InitSession(sessionKey string) (*AppSession, bool) {
	sess, ok := m.SessionContainer.InitSession(sessionKey, NewSession())
	if !ok {
		return nil, false
	}

	mathSess, ok := sess.(*AppSession)
	if !ok {
		return mathSess, false
	}

	return mathSess, true
}

func (m *SessionManager) DeleteSession(sessionKey string) {
	m.SessionContainer.DeleteSession(sessionKey)
}

func (m *SessionManager) MarkExpiredSessions() {
	for _, sess := range *m.Sessions() {
		key, ok := sess.Get("key")
		if !ok {
			continue
		}

		if sess.IsExpired() {
			log.Printf("MarkExpiredSessions: marked session %s for deletion", key.(string))
			sess.MarkForDeletion()
		}
	}
}

func (m *SessionManager) DeleteExpiredSessions() {
	for _, sess := range *m.Sessions() {
		key, ok := sess.Get("key")
		if !ok {
			continue
		}

		if sess.IsExpired() {
			log.Printf("DeleteExpiredSessions: deleting session %s", key.(string))
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
}
