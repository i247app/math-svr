package enum

const (
	RoleAdmin ERole = "admin"
	RoleUser  ERole = "user"
	RoleGuest ERole = "guest"

	StatusActive   EStatus = "ACTIVE"
	StatusInactive EStatus = "INACTIVE"
	StatusBanned   EStatus = "BANNED"

	LoginStatusActive            ELoginStatus = "ACTIVE"
	LoginStatusTwoFactorRequired ELoginStatus = "2FA_REQUIRED"

	DefaultLang = "en"
)
