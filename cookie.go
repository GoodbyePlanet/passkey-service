package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func isProduction() bool {
	return os.Getenv("ENV") == "production"
}

func SetCookie(c *gin.Context, name, value, path string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   isProduction(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearCookie(c *gin.Context, name, path string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Secure:   isProduction(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
