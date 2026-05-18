package bootstrap

import (
	"fmt"
	"log"
	"os"

	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/utils"
)

func (a *App) SerializeSessions(filename string) error {
	sessDump := session.Dump(a.Resource.SessionManager)
	sessionsData := make(map[string]any)
	for k, v := range *sessDump {
		log.Printf("Serializing session: %s", truncateSessionKey(k))
		sessionsData[k] = v
	}

	data, err := utils.SerializeMap(&sessionsData)
	if err != nil {
		return fmt.Errorf("failed to serialize session manager: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to session manager file: %w", err)
	}

	return nil
}

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

		log.Printf("Reloading session: %s", truncateSessionKey(sessionKey))
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

func truncateSessionKey(key string) string {
	// If longer than 19 characters, shorten the key to the first and last 8 characters
	if len(key) > 19 {
		return key[:8] + "..." + key[len(key)-8:]
	}
	return key
}
