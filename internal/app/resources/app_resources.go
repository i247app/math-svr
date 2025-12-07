package resources

import (
	"fmt"
	"net/http"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/db"
)

type AppResource struct {
	Env            *config.Env
	HostConfig     gex.HostConfig
	Db             db.IDatabase
	SessionManager *session.SessionManager
}

func (a *AppResource) GetRequestSession(r *http.Request) (*session.AppSession, error) {
	sess := session.GetRequestSession(r)
	if sess == nil {
		return nil, fmt.Errorf("session not found")
	}

	return sess, nil
}

func (a *AppResource) GetRequestUID(r *http.Request) (string, error) {
	sess, err := a.GetRequestSession(r)
	if err != nil {
		return "", err
	}

	id, ok := sess.UID()
	if !ok {
		return "", fmt.Errorf("uid missing from session (did you forget to send the Authorization header?)")
	}

	return id, nil
}
