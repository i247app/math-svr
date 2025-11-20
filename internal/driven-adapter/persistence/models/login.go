package models

import "time"

type LoginModel struct {
	ID        string
	UID       string
	HashPass  string
	Status    string
	CreateDT  time.Time
	ModifyDT  time.Time
	DeletedDT *time.Time
}

type LoginLogModel struct {
	ID         string
	UID        string
	IPaddress  string
	DeviceUUID string
	Token      string
	Status     string
	CreateDT   time.Time
	ModifyDT   time.Time
	DeletedDT  *time.Time
}
