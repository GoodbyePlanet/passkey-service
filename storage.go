package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"passkey-service/config"
	"passkey-service/models"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

func FindOrCreateUser(username string, displayName string) (*models.User, error) {
	user := models.User{Username: username, DisplayName: displayName}
	if err := config.DB.FirstOrCreate(&user, models.User{Username: username}).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string) *models.User {
	user := models.User{}
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil
	}
	return &user
}

func GetUserWithCredentials(username string) (*models.User, error) {
	user := models.User{}
	if err := config.DB.Preload("Credentials").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateOrUpdateCredential(cred *webauthn.Credential, user *models.User) (*models.Credential, error) {
	dbCred := models.Credential{
		UserID:          user.ID,
		ID:              cred.ID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		Transport:       models.TransportToStrings(cred.Transport),

		UserPresent:    cred.Flags.UserPresent,
		UserVerified:   cred.Flags.UserVerified,
		BackupEligible: cred.Flags.BackupEligible,
		BackupState:    cred.Flags.BackupState,

		AAGUID:       cred.Authenticator.AAGUID,
		SignCount:    cred.Authenticator.SignCount,
		CloneWarning: cred.Authenticator.CloneWarning,
		Attachment:   string(cred.Authenticator.Attachment),

		ClientDataJSON:     cred.Attestation.ClientDataJSON,
		ClientDataHash:     cred.Attestation.ClientDataHash,
		AuthenticatorData:  cred.Attestation.AuthenticatorData,
		Object:             cred.Attestation.Object,
		PublicKeyAlgorithm: cred.Attestation.PublicKeyAlgorithm,
	}

	var existing models.Credential
	err := config.DB.Where("id = ?", cred.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := config.DB.Create(&dbCred).Error; err != nil {
				return nil, err
			}
			return &dbCred, nil
		}
		return nil, err
	}

	// Update existing record
	dbCred.ID = existing.ID
	if err := config.DB.Model(&existing).Updates(dbCred).Error; err != nil {
		return nil, err
	}
	return &dbCred, nil
}

func SaveWebAuthnSession(session *webauthn.SessionData, username string) (string, error) {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", err
	}
	sessionID := generateSessionID()
	err = config.DB.Save(&models.WebAuthnSession{
		SessionID:  sessionID,
		Username:   username,
		SessionRaw: sessionJSON,
	}).Error
	return sessionID, err
}

func GetWebAuthnSession(sessionID string) (*models.WebAuthnSession, error) {
	var was models.WebAuthnSession
	if err := config.DB.Where("session_id = ?", sessionID).First(&was).Error; err != nil {
		return nil, err
	}
	return &was, nil
}

func RemoveWebAuthnSession(sessionID string) error {
	return config.DB.Delete(&models.WebAuthnSession{}, "session_id = ?", sessionID).Error
}

func generateSessionID() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("failed to generate session ID:")
	}

	return base64.URLEncoding.EncodeToString(b)
}
