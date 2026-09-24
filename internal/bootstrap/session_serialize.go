package bootstrap

import (
	"fmt"
	"log"
	"os"

	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/utils"
)

func (a *App) ReloadSessions(filename string) error {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("no session file [%s] found", filename)
	}

	dump, err := utils.DeserializeMap(fileData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session manager file: %w", err)
	}

	log.Printf("Loaded %d sessions from %s\n", len(*dump), filename)

	for sessionKey, sessionData := range *dump {
		structuredData, ok := sessionData.(map[string]any)
		if !ok {
			log.Printf("Session data for %s is not a map[string]interface{}\n", sessionKey)
			continue
		}

		// Sessions are keyed by their signed token (gex JwtSessionProvider).
		// Files written before that change keyed them by the claim's
		// session_key; re-key those by their token so their holders stay
		// signed in across the upgrade.
		if token, ok := structuredData["token"].(string); ok && token != "" && token != sessionKey {
			sessionKey = token
			structuredData["key"] = token
		}

		log.Printf("Reloading session: %s", session.ShortKey(sessionKey))
		sess, ok := a.Resource.SessionManager.InitSession(sessionKey)
		if ok {
			for k, v := range structuredData {
				sess.Put(k, v)
			}
		} else {
			log.Println("* error")
			// return fmt.Errorf("failed to create session")
		}
	}

	return nil
}
