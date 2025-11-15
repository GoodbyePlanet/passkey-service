package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn

type RegistrationBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type LoginBeginRequest struct {
	Username string `json:"username"`
}

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

	user := FindOrCreateUser(username, displayName)
	options, session, err := webAuthn.BeginRegistration(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := SaveWebAuthnSession(session, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, options)
}

// FinishRegistration POST /api/register/finish
func FinishRegistration(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
		return
	}
	user := GetUserByUsername(username)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	session, errParsingSession := getAndParseWebAuthnSession(username)
	if errParsingSession != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errParsingSession.Error()})
		return
	}

	cred, errFinishReg := webAuthn.FinishRegistration(user, *session, c.Request)
	if errFinishReg != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errFinishReg.Error()})
		return
	}

	_, errCreateCred := CreateCredential(cred, user)
	if errCreateCred != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errCreateCred})
		return
	}

	if err := RemoveWebAuthnSession(username).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// BeginLogin POST /api/authenticate/begin
func BeginLogin(c *gin.Context) {
	var req LoginBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	username := req.Username
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
		return
	}

	user := GetUserByUsername(username)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	options, session, err := webAuthn.BeginLogin(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := SaveWebAuthnSession(session, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, options)
}

// FinishLogin POST /api/authenticate/finish
func FinishLogin(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
		return
	}
	user := GetUserByUsername(username)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	session, errParsingSession := getAndParseWebAuthnSession(username)
	if errParsingSession != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errParsingSession.Error()})
		return
	}

	_, errFinishLogin := webAuthn.FinishLogin(user, *session, c.Request)
	if errFinishLogin != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errFinishLogin.Error()})
		return
	}

	if err := RemoveWebAuthnSession(username).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "authenticated"})
}

func getAndParseWebAuthnSession(username string) (*webauthn.SessionData, error) {
	was, err := GetWebAuthnSession(username)
	if err != nil {
		return nil, err
	}

	var sd webauthn.SessionData
	if err := json.Unmarshal(was.SessionRaw, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}
