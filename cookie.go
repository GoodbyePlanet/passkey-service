package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetCookie(c *gin.Context, name, value, path string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   IsProduction(),
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
		Secure:   IsProduction(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
