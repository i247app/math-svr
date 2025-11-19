package session

import (
	"time"

	"maps"

	"github.com/i247app/gex/session"
)

const (
	DefaultSessionTTL = time.Second * 10 // 10 seconds
)

type AppSession struct {
	GexSession *session.InMemorySession
}

func NewSession() *AppSession {
	gexSess := session.NewInMemorySession()
	return &AppSession{GexSession: gexSess}
}

func (s *AppSession) Put(key string, value any) {
	s.GexSession.Put(key, value)
}

func (s *AppSession) Get(key string) (any, bool) {
	return s.GexSession.Get(key)
}

func (s *AppSession) ToMap() map[string]any {
	result := make(map[string]any)
	maps.Copy(result, s.GexSession.Data)
	return result
}

func (s *AppSession) UID() (int64, bool) {
	result, ok := s.Get("uid")
	if !ok {
		// log.Println("ERROR: key 'uid' not in session store")
		return 0, false
	}

	uid, ok := result.(int64)
	if !ok {
		// log.Printf("ERROR: 'uid' is in session store but expected to be an int64, is a %T\n", result)
		return 0, false
	}

	return uid, true
}

func (s *AppSession) MarkForDeletion() {
	s.Put("marked_for_deletion", true)

	// TODO: maybe do this instead
	// s.Put("expires_at", time.Now().Add(-time.Minute*10))
}

type InitData struct {
	Source    string
	IsSecure  bool
	UID       int64
	Email     string
	LoginName string
	ExpireAt  *time.Time
}

func (s *AppSession) Init(data InitData) *AppSession {
	var expireAt time.Time
	if data.ExpireAt == nil {
		expireAt = time.Now().Add(DefaultSessionTTL)
	} else {
		expireAt = *data.ExpireAt
	}

	s.Put("source", data.Source)
	s.Put("is_secure", data.IsSecure)
	s.Put("uid", data.UID)
	s.Put("email", data.Email)
	s.Put("login_name", data.LoginName)
	s.Put("expires_at", expireAt)
	return s
}

func (s *AppSession) IsExpired() bool {
	expireAt, ok := s.Get("expires_at")
	if !ok {
		return false
	}

	return time.Now().After(expireAt.(time.Time))
}

func (s *AppSession) DeviceID() (int64, bool) {
	device, ok := s.Get("device")
	if !ok {
		return 0, false
	}
	deviceID := device.(map[string]any)["device_id"]
	return deviceID.(int64), true
}

func (s *AppSession) MarkNotSecure() {
	s.Put("is_secure", false)
}

func (s *AppSession) MarkExpired() {
	s.Put("expires_at", time.Now().Add(-time.Minute*10))
}

func (s *AppSession) IsMarkedForDeletion() bool {
	markedForDeletion, ok := s.Get("marked_for_deletion")
	return ok && markedForDeletion.(bool)
}

func (s *AppSession) IsValid() bool {
	return !s.IsExpired() && !s.IsMarkedForDeletion()
}
