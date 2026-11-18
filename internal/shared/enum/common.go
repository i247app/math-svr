package enum

import "strings"

type LanguageType string

const (
	LanguageTypeVietnamese   LanguageType = "vi"
	LanguageTypeVietnameseV2 LanguageType = "vi-VN"
	LanguageTypeEnglish      LanguageType = "en"
	LanguageTypeEnglishV2    LanguageType = "en-US"
)

func (s LanguageType) String() string {
	return string(s)
}

func (s LanguageType) IsValid() bool {
	switch s {
	case LanguageTypeVietnamese, LanguageTypeVietnameseV2,
		LanguageTypeEnglish, LanguageTypeEnglishV2:
		return true
	default:
		return false
	}
}

// ContentLanguages are the languages question content is written in (a
// question's base language and the keys of its translations): the short
// codes only.
var ContentLanguages = []LanguageType{LanguageTypeVietnamese, LanguageTypeEnglish}

// Base is the short code of a known language ("vi-VN" → "vi", " EN " →
// "en"), "" for anything else.
func (s LanguageType) Base() LanguageType {
	switch LanguageType(strings.ToLower(strings.TrimSpace(string(s)))) {
	case LanguageTypeVietnamese, "vi-vn":
		return LanguageTypeVietnamese
	case LanguageTypeEnglish, "en-us":
		return LanguageTypeEnglish
	}
	return ""
}

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
