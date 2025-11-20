package enum

type ERole string
type EStatus string

const (
	RoleAdmin ERole = "admin"
	RoleUser  ERole = "user"
	RoleGuest ERole = "guest"

	StatusActive   EStatus = "ACTIVE"
	StatusInactive EStatus = "INACTIVE"
	StatusBanned   EStatus = "BANNED"

	DefaultLang = "en"
)
