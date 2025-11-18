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
		RPOrigins:     GetEnvList("RP_ORIGINS"),
	})
	if err != nil {
		panic("failed to init webauthn: " + err.Error())
	}
}

// BeginRegistration POST /api/register/begin
func BeginRegistration(c *gin.Context) {
	logger.Info("BeginRegistration flow started...")
	var req RegistrationBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithInvalidJSON(c)
		return
	}
	username := req.Username
	displayName := req.DisplayName

	user, errCreatingUser := FindOrCreateUser(username, displayName)
	if errCreatingUser != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to create user: "+errCreatingUser.Error())
		return
	}

	options, session, err := webAuthn.BeginRegistration(user)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to begin registration: "+err.Error())
		return
	}

	sessionID, err := SaveWebAuthnSession(session, username)
	if err != nil {
		respondWithFailedToSaveSession(c)
		return
	}

	SetCookie(c, "sid", sessionID, "/api/register", 3600)
	c.JSON(http.StatusOK, options)
}

// FinishRegistration POST /api/register/finish
func FinishRegistration(c *gin.Context) {
	logger.Info("FinishRegistration flow started...")
	sid, err := c.Cookie("sid")
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "failed to get session cookie")
	}

	session, errParsingSession := getAndParseWebAuthnSession(sid)
	if errParsingSession != nil {
		respondWithFailedToParseSession(c)
		return
	}

	user := GetUserByUsername(string(session.UserID))
	if user == nil {
		respondWithUserNotFound(c)
		return
	}

	cred, errFinishReg := webAuthn.FinishRegistration(user, *session, c.Request)
	if errFinishReg != nil {
		respondWithError(c, http.StatusBadRequest, "failed to finish registration: "+errFinishReg.Error())
		return
	}

	_, errCreateCred := CreateOrUpdateCredential(cred, user)
	if errCreateCred != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to create credential: "+errCreateCred.Error())
		return
	}

	if err := RemoveWebAuthnSession(sid); err != nil {
		respondWithFailedToRemoveSession(c)
		return
	}

	ClearCookie(c, "sid", "/api/register")
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// BeginLogin POST /api/authenticate/begin
func BeginLogin(c *gin.Context) {
	logger.Info("BeginLogin flow started...")
	var req LoginBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithInvalidJSON(c)
		return
	}
	username := req.Username
	if username == "" {
		responseWithMissingUsername(c)
		return
	}

	user, errUserCredentials := GetUserWithCredentials(username)
	if errUserCredentials != nil {
		respondWithUserNotFound(c)
		return
	}

	options, session, errBeginLogin := webAuthn.BeginLogin(user)
	if errBeginLogin != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to begin login: "+errBeginLogin.Error())
		return
	}

	sessionID, err := SaveWebAuthnSession(session, username)
	if err != nil {
		respondWithFailedToSaveSession(c)
		return
	}

	SetCookie(c, "sid", sessionID, "/api/authenticate", 3600)
	c.JSON(http.StatusOK, options)
}

// FinishLogin POST /api/authenticate/finish
func FinishLogin(c *gin.Context) {
	logger.Info("FinishLogin flow started...")
	sid, err := c.Cookie("sid")
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "failed to get session cookie")
	}

	session, errParsingSession := getAndParseWebAuthnSession(sid)
	if errParsingSession != nil {
		respondWithFailedToParseSession(c)
		return
	}

	user, errUserCredentials := GetUserWithCredentials(string(session.UserID))
	if errUserCredentials != nil {
		respondWithUserNotFound(c)
		return
	}

	cred, errFinishLogin := webAuthn.FinishLogin(user, *session, c.Request)
	if errFinishLogin != nil {
		respondWithError(c, http.StatusUnauthorized, "failed to finish login: "+errFinishLogin.Error())
		return
	}

	_, errUpdateCred := CreateOrUpdateCredential(cred, user)
	if errUpdateCred != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to update credential: "+errUpdateCred.Error())
		return
	}

	if err := RemoveWebAuthnSession(sid); err != nil {
		respondWithFailedToRemoveSession(c)
		return
	}

	ClearCookie(c, "sid", "/api/authenticate")
	c.JSON(http.StatusOK, gin.H{"status": "authenticated"})
}

func getAndParseWebAuthnSession(sid string) (*webauthn.SessionData, error) {
	was, err := GetWebAuthnSession(sid)
	if err != nil {
		return nil, err
	}

	var sd webauthn.SessionData
	if err := json.Unmarshal(was.SessionRaw, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func respondWithInvalidJSON(c *gin.Context) {
	respondWithError(c, http.StatusBadRequest, "invalid JSON")
}

func responseWithMissingUsername(c *gin.Context) {
	respondWithError(c, http.StatusBadRequest, "missing username")
}

func respondWithUserNotFound(c *gin.Context) {
	respondWithError(c, http.StatusNotFound, "user not found")
}

func respondWithFailedToParseSession(c *gin.Context) {
	respondWithError(c, http.StatusBadRequest, "failed to parse session")
}

func respondWithFailedToRemoveSession(c *gin.Context) {
	respondWithError(c, http.StatusInternalServerError, "failed to remove session")
}

func respondWithFailedToSaveSession(c *gin.Context) {
	respondWithError(c, http.StatusInternalServerError, "failed to save session")
}
