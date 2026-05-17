package status

type MStatus struct {
	code    StatusCode
	message StatusMessage
}

func NewMStatus(code StatusCode, message StatusMessage) *MStatus {
	return &MStatus{
		code:    code,
		message: message,
	}
}

func (s *MStatus) Code() StatusCode {
	return s.code
}

func (s *MStatus) Message() StatusMessage {
	return s.message
}
