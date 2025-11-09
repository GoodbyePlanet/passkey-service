package main

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register/begin", BeginRegistration)
		api.POST("/register/finish", FinishRegistration)
		api.POST("/authenticate/begin", BeginLogin)
		api.POST("/authenticate/finish", FinishLogin)
	}

	return r
}
