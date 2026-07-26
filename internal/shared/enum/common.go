package enum

type LanguageType string

const (
	LanguageTypeVietnamese   LanguageType = "vi"
	LanguageTypeVietnameseV2 LanguageType = "vi-VN"
	LanguageTypeEnglish      LanguageType = "en"
	LanguageTypeEnglishV2    LanguageType = "en-US"
)

type StatusType string

const (
	StatusActive   StatusType = "ACTIVE"
	StatusInactive StatusType = "INACTIVE"
)

func (s StatusType) String() string {
	return string(s)
}

func (s StatusType) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive:
		return true
	default:
		return false
	}
}
