package main

import (
	"encoding/json"
	"passkey-service/config"
	"passkey-service/models"

	"github.com/go-webauthn/webauthn/webauthn"
)

func FindOrCreateUser(username string, displayName string) *models.User {
	user := models.User{Username: username, DisplayName: displayName}
	config.DB.FirstOrCreate(&user, models.User{Username: username})
	return &user
}

func GetUserByUsername(username string) *models.User {
	user := models.User{}
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil
	}
	return &user
}

func CreateCredential(cred *webauthn.Credential, user *models.User) (*models.Credential, error) {
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

	if err := config.DB.Create(&dbCred).Error; err != nil {
		return nil, err
	}
	return &dbCred, nil
}

func SaveWebAuthnSession(session *webauthn.SessionData, username string) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return config.DB.Save(&models.WebAuthnSession{
		Username:   username,
		SessionRaw: sessionJSON,
	}).Error
}

func GetWebAuthnSession(username string) (*models.WebAuthnSession, error) {
	var was models.WebAuthnSession
	if err := config.DB.Where("username = ?", username).First(&was).Error; err != nil {
		return nil, err
	}
	return &was, nil
}

func RemoveWebAuthnSession(username string) error {
	return config.DB.Delete(&models.WebAuthnSession{}, "username = ?", username).Error
}
