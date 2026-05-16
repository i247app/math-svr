package utils

import "github.com/google/uuid"

func GenerateUUID() uuid.UUID {
	return uuid.New()
}

func StringToUUID(uuidStr string) uuid.UUID {
	return uuid.Must(uuid.Parse(uuidStr))
}
