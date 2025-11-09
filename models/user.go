package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex"`
	DisplayName string
	Credentials []Credential `gorm:"foreignKey:UserID"`
}

type Credential struct {
	gorm.Model
	UserID      uint
	ID          []byte
	PublicKey   []byte
	Attestation string
	SignCount   uint32
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.Username)
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = webauthn.Credential{
			ID:        c.ID,
			PublicKey: c.PublicKey,
			Authenticator: webauthn.Authenticator{
				AAGUID:    make([]byte, 16),
				SignCount: c.SignCount,
			},
		}
	}
	return creds
}
