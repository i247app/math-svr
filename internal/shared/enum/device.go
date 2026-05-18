package enum

type DeviceStatusType string

const (
	DeviceStatusTypeActive   DeviceStatusType = "ACTIVE"
	DeviceStatusTypeInactive DeviceStatusType = "INACTIVE"
	DeviceStatusTypeRevoked  DeviceStatusType = "REVOKED"
	DeviceStatusTypeDeleted  DeviceStatusType = "DELETED"
)

func (s DeviceStatusType) String() string {
	return string(s)
}

func (s DeviceStatusType) IsValid() bool {
	switch s {
	case DeviceStatusTypeActive, DeviceStatusTypeInactive, DeviceStatusTypeRevoked:
		return true
	default:
		return false
	}
}
