package models

import (
	"github.com/go-webauthn/webauthn/protocol"
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

	UserID uint

	ID        []byte
	PublicKey []byte

	// Metadata
	AttestationType string
	Transport       []string

	// Flags
	UserPresent    bool
	UserVerified   bool
	BackupEligible bool
	BackupState    bool

	// Authenticator
	AAGUID       []byte
	SignCount    uint32
	CloneWarning bool
	Attachment   string

	// Attestation
	ClientDataJSON     []byte
	ClientDataHash     []byte
	AuthenticatorData  []byte
	Object             []byte
	PublicKeyAlgorithm int64
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

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))

	for i, c := range u.Credentials {
		creds[i] = webauthn.Credential{
			ID:        c.ID,
			PublicKey: c.PublicKey,

			AttestationType: c.AttestationType,

			Transport: toTransports(c.Transport),

			Flags: webauthn.CredentialFlags{
				UserPresent:    c.UserPresent,
				UserVerified:   c.UserVerified,
				BackupEligible: c.BackupEligible,
				BackupState:    c.BackupState,
			},

			Authenticator: webauthn.Authenticator{
				AAGUID:       c.AAGUID,
				SignCount:    c.SignCount,
				CloneWarning: c.CloneWarning,
				Attachment:   protocol.AuthenticatorAttachment(c.Attachment),
			},

			Attestation: webauthn.CredentialAttestation{
				ClientDataJSON:     c.ClientDataJSON,
				ClientDataHash:     c.ClientDataHash,
				AuthenticatorData:  c.AuthenticatorData,
				Object:             c.Object,
				PublicKeyAlgorithm: c.PublicKeyAlgorithm,
			},
		}
	}
	return creds
}

func TransportToStrings(t []protocol.AuthenticatorTransport) []string {
	out := make([]string, len(t))
	for i, v := range t {
		out[i] = string(v)
	}
	return out
}

func toTransports(strs []string) []protocol.AuthenticatorTransport {
	t := make([]protocol.AuthenticatorTransport, 0, len(strs))
	for _, s := range strs {
		t = append(t, protocol.AuthenticatorTransport(s))
	}
	return t
}
