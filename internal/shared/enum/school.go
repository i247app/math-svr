package enum

type SchoolStatusType string

const (
	SchoolStatusTypeActive   SchoolStatusType = "ACTIVE"
	SchoolStatusTypeInactive SchoolStatusType = "INACTIVE"
	SchoolStatusTypeDeleted  SchoolStatusType = "DELETED"
)

func (s SchoolStatusType) String() string {
	return string(s)
}

func (s SchoolStatusType) IsValid() bool {
	switch s {
	case SchoolStatusTypeActive, SchoolStatusTypeInactive:
		return true
	default:
		return false
	}
}
