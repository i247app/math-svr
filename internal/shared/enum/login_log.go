package enum

type LoginLogStatusType string

const (
	LoginLogStatusTypeActive   LoginLogStatusType = "ACTIVE"
	LoginLogStatusTypeInactive LoginLogStatusType = "INACTIVE"
	LoginLogStatusTypeRevoked  LoginLogStatusType = "REVOKED"
	LoginLogStatusTypeDeleted  LoginLogStatusType = "DELETED"
)

func (s LoginLogStatusType) String() string {
	return string(s)
}

func (s LoginLogStatusType) IsValid() bool {
	switch s {
	case LoginLogStatusTypeActive, LoginLogStatusTypeInactive, LoginLogStatusTypeRevoked:
		return true
	default:
		return false
	}
}
