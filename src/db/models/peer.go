package models

import (
	"database/sql/driver"

	"gorm.io/gorm"
)

type Peer struct {
	gorm.Model

	IpAddress     string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	TrafficAmount int
	Status        Status
}

type Status int64

const (
	Unused Status = iota
	Enabled
	Disabled
)

func (u *Status) Scan(value interface{}) error { *u = Status(value.(int64)); return nil }
func (u Status) Value() (driver.Value, error)  { return int64(u), nil }
