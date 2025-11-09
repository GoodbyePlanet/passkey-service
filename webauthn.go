package main

import (
	"log/slog"
	"net/http"
	"os"
	"passkey-service/config"
	"passkey-service/models"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn

// TODO: Store sessions in redis or in postgres
var sessionStore = map[string]*webauthn.SessionData{}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func InitWebAuthn() {
	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Passkey Service",
		RPID:          os.Getenv("RP_ID"),
		RPOrigins:     []string{os.Getenv("RP_ORIGIN")},
	})
	if err != nil {
		panic("failed to init webauthn: " + err.Error())
	}
}

// BeginRegistration POST /api/register/begin
func BeginRegistration(c *gin.Context) {
	username := c.PostForm("username")
	displayName := c.PostForm("displayName")

	logger.Info("register begin ", username, displayName)

	user := models.User{Username: username, DisplayName: displayName}
	config.DB.FirstOrCreate(&user, models.User{Username: username})

	options, session, err := webAuthn.BeginRegistration(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sessionStore[username] = session
	c.JSON(http.StatusOK, options)
}

// FinishRegistration POST /api/register/finish
func FinishRegistration(c *gin.Context) {
	username := c.PostForm("username")
	user := models.User{}
	config.DB.Where("username = ?", username).First(&user)

	session := sessionStore[username]
	credential, err := webAuthn.FinishRegistration(&user, *session, c.Request)
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
