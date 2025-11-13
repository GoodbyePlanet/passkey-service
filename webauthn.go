package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"passkey-service/config"
	"passkey-service/models"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn

type RegistrationBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// TODO: Store sessions in redis or in postgres
var sessionStore = map[string]*webauthn.SessionData{}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func InitWebAuthn() {
	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Passkey Service",
		RPID:          os.Getenv("RP_ID"),
		RPOrigins:     []string{"http://localhost:63342", "http://localhost:8080"},
	})
	if err != nil {
		panic("failed to init webauthn: " + err.Error())
	}
}

// BeginRegistration POST /api/register/begin
func BeginRegistration(c *gin.Context) {
	var req RegistrationBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	username := req.Username
	displayName := req.DisplayName
	logger.Info("register begin for user ", username, displayName)

	user := models.User{Username: username, DisplayName: displayName}
	config.DB.FirstOrCreate(&user, models.User{Username: username})

	options, session, err := webAuthn.BeginRegistration(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sessionJSON, _ := json.Marshal(session)
	config.DB.Save(&models.WebAuthnSession{
		Username:   username,
		SessionRaw: sessionJSON,
	})
	c.JSON(http.StatusOK, options)
}

// FinishRegistration POST /api/register/finish
func FinishRegistration(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
		return
	}
	user := models.User{}
	config.DB.Where("username = ?", username).First(&user)

	credential, err := webAuthn.FinishRegistration(&user, getSession(c, username), c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&models.Credential{
		UserID:    user.ID,
		ID:        credential.ID,
		PublicKey: credential.PublicKey,
		SignCount: credential.Authenticator.SignCount,
	})
	config.DB.Delete(&models.WebAuthnSession{}, "username = ?", username)

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// BeginLogin POST /api/authenticate/begin
func BeginLogin(c *gin.Context) {
	username := c.PostForm("username")
	user := models.User{}
	if err := config.DB.Preload("Credentials").Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	options, session, err := webAuthn.BeginLogin(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionStore[username] = session
	c.JSON(http.StatusOK, options)
}

// FinishLogin POST /api/authenticate/finish
func FinishLogin(c *gin.Context) {
	username := c.PostForm("username")
	user := models.User{}
	config.DB.Preload("Credentials").Where("username = ?", username).First(&user)

	session := sessionStore[username]
	_, err := webAuthn.FinishLogin(&user, *session, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "authenticated"})
}

func getSession(c *gin.Context, username string) webauthn.SessionData {
	var was models.WebAuthnSession
	if err := config.DB.Where("username = ?", username).First(&was).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
	}

	var sd webauthn.SessionData
	if err := json.Unmarshal(was.SessionRaw, &sd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session data"})
	}

	return sd
}
