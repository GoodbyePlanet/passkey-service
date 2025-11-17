package models

import (
	"time"

	"gorm.io/datatypes"
)

type WebAuthnSession struct {
	SessionID  string `gorm:"primaryKey"`
	Username   string
	SessionRaw datatypes.JSON `gorm:"column:session_data"`
	CreatedAt  time.Time
}
