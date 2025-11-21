package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	config := cors.Config{
		AllowOrigins:     GetEnvList("RP_ORIGINS"),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
	r.Use(cors.New(config))

	api := r.Group("/api")
	{
		api.POST("/register/begin", BeginRegistration)
		api.POST("/register/finish", FinishRegistration)
		api.POST("/authenticate/begin", BeginLogin)
		api.POST("/authenticate/finish", FinishLogin)
		api.GET("/users/:username/registered-passkeys", GetRegisteredPasskeys)
	}

	return r
}
