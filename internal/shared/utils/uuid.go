package utils

import "github.com/google/uuid"

func GenerateUUID() uuid.UUID {
	return uuid.New()
}

func StringToUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

func PtrStringToUUID(uuidStr *string) (uuid.UUID, error) {
	if uuidStr == nil {
		return uuid.Nil, nil
	}

	return uuid.Parse(*uuidStr)
}

func IsEmptyUUID(u uuid.UUID) bool {
	return u == uuid.Nil
}
