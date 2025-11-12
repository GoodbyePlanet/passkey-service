package models

import (
	"time"

	"gorm.io/datatypes"
)

type WebAuthnSession struct {
	Username   string         `gorm:"primaryKey"`
	SessionRaw datatypes.JSON `gorm:"column:session_data"`
	CreatedAt  time.Time
}
