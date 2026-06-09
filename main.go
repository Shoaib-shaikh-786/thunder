package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backend/internal/order"
	"backend/internal/pkg/auth"
	"backend/internal/pkg/config"
	"backend/internal/pkg/db"
	"backend/internal/pkg/log"
	"backend/internal/product"
	"backend/internal/tenant"
	"backend/internal/user"
)

var (
	cfg         *config.Config
	authService auth.BaseAuthService
)

func main() {
	ctx := context.Background()

	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.ErrorWithContext(ctx, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	pool, err := db.NewPool(ctx, cfg.DBConfig.DSN)
	if err != nil {
		log.ErrorWithContext(ctx, fmt.Sprintf("failed to connect to database: %v", err))
		return
	}
	defer pool.Close()

	// User layer
	userRepo := user.NewRepository(pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Product layer
	productRepo := product.NewRepository(pool)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// Order layer
	orderRepo := order.NewRepository(pool)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Tenant layer
	tenantRepo := tenant.NewRepository(pool)
	tenantService := tenant.NewService(tenantRepo)
	tenantHandler := tenant.NewHandler(tenantService)

	// Auth middleware — wraps session token validation
	authService = auth.NewSessionAuthService(func(ctx context.Context, token string) (*auth.Claims, error) {
		c, err := userService.ValidateToken(ctx, token)
		if err != nil {
			return nil, err
		}
		return c, nil
	})

	router := initialiseRouter(userHandler, productHandler, orderHandler, tenantHandler)
	if err := router.Run(fmt.Sprintf("%s:%s", cfg.ApplicationConfig.ServerHost, cfg.ApplicationConfig.ServerPort)); err != nil {
		log.ErrorWithContext(ctx, fmt.Sprintf("failed to start server: %v", err))
	}
}

func initialiseRouter(
	userHandler *user.Handler,
	productHandler *product.Handler,
	orderHandler *order.Handler,
	tenantHandler *tenant.Handler,
) *gin.Engine {
	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowMethods = []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	r.Use(cors.New(corsConfig))

	r.Use(auth.AuthMiddleware(authService))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	userHandler.RegisterRoutes(api)
	productHandler.RegisterRoutes(api)
	orderHandler.RegisterRoutes(api)
	tenantHandler.RegisterRoutes(api)

	return r
}
