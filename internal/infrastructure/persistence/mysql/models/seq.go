package models

import "time"

type SeqModel struct {
	SeqName      string
	CurrentValue uint64
	Prefix       string
	Padding      uint8
	IncrementBy  uint32
	ModifyDt     time.Time
}
