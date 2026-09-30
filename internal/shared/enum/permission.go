package enum

// RoleRecordStatusType is the business lifecycle state of a row in the
// permission module's role registry (ma_roles.role_status). It is unrelated
// to RoleType, which is the free-text role carried on ma_users / ma_profiles.
// DELETED marks a soft delete and is never accepted from a client.
type RoleRecordStatusType string

const (
	RoleRecordStatusActive   RoleRecordStatusType = "ACTIVE"
	RoleRecordStatusInactive RoleRecordStatusType = "INACTIVE"
	RoleRecordStatusDeleted  RoleRecordStatusType = "DELETED"
)

func (s RoleRecordStatusType) String() string {
	return string(s)
}

func (s RoleRecordStatusType) IsValid() bool {
	switch s {
	case RoleRecordStatusActive, RoleRecordStatusInactive:
		return true
	default:
		return false
	}
}
