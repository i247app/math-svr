package status

type KStatus struct {
	code    StatusCode
	message StatusMessage
}

func NewKStatus(code StatusCode, message StatusMessage) *KStatus {
	return &KStatus{
		code:    code,
		message: message,
	}
}

func (s *KStatus) Code() StatusCode {
	return s.code
}

func (s *KStatus) Message() StatusMessage {
	return s.message
}

// func (s *KStatus) SetMessage(language enum.LanguageType, code StatusCode, args map[string]any) {
// 	s.message = GetMessage(language, code, args)
// }
