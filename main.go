package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backend/internal/pkg/auth"
	"backend/internal/pkg/auth/jwt"
	"backend/internal/pkg/config"
	"backend/internal/pkg/log"
)

var (
	cfg         *config.Config
	authService auth.BaseAuthService
)

func main() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.ErrorWithContext(context.Background(), fmt.Sprintf("failed to load configuration: %v", err))
		return
	}

	authService, err = initialiseAuthService(cfg)
	if err != nil {
		log.ErrorWithContext(context.Background(), fmt.Sprintf("failed to initialise auth service: %v", err))
		return
	}

	router := initialiseRouter()
	if err := router.Run(fmt.Sprintf("%s:%s", cfg.ApplicationConfig.ServerHost, cfg.ApplicationConfig.ServerPort)); err != nil {
		log.ErrorWithContext(context.Background(), fmt.Sprintf("failed to start server: %v", err))
	}
}

func initialiseAuthService(cfg *config.Config) (auth.BaseAuthService, error) {
	return jwt.NewJWTAuth(cfg)
}

func initialiseRouter() *gin.Engine {
	r := gin.Default()

	// Cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

	r.Use(cors.New(corsConfig))

	r.Use(auth.AuthMiddleware(authService))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
