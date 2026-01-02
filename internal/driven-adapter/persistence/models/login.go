package models

import "time"

type LoginModel struct {
	ID          string
	UID         string
	HashPass    string
	Note        *string
	LoginStatus string
	Status      string
	CreateID    *int64
	CreateDT    time.Time
	ModifyID    *int64
	ModifyDT    time.Time
	DeletedDT   *time.Time
}

type LoginLogModel struct {
	ID             string
	UID            string
	IPaddress      string
	DeviceUUID     string
	Token          string
	Note           *string
	LoginLogStatus string
	Status         string
	CreateID       *int64
	CreateDT       time.Time
	ModifyID       *int64
	ModifyDT       time.Time
	DeletedDT      *time.Time
}
