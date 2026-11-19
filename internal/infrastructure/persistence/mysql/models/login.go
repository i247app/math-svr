package models

type LoginModel struct {
	Audit

	LoginId      int64
	Uid          int64
	Upw          *string
	LoginsStatus *string
}
