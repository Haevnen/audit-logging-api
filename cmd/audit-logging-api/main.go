package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/config"
	"github.com/Haevnen/audit-logging-api/internal/infra/middleware"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/pkg/gormdb"
	"github.com/Haevnen/audit-logging-api/pkg/logger"
)

func registerHandlersWithOptionsForTenant(router gin.IRouter, si api_service.ServerInterface, options api_service.GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := api_service.ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.GET(options.BaseURL+"/tenants", wrapper.ListTenants)
	router.POST(options.BaseURL+"/tenants", wrapper.CreateTenant)
}

func start() int {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger := logger.InitLogger(cfg.Mode, cfg.LogFile)

	// Connect to database
	gormCfg := cfg.GetGORMConfig()
	db, closeFunc, err := gormdb.New(gormCfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to connect to database")
		return 1
	}
	defer func() {
		if err := closeFunc(); err != nil {
			logger.WithField("error", err).Fatal("Failed to close connection to database")
		}
	}()

	// Setup router
	r := gin.Default()
	gin.SetMode(cfg.Mode)

	registry := registry.NewRegistry(db, cfg.TokenSymmetricKey)
	handler := handler.New(registry)
	jwt := registry.Manager()

	// Register health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/v1/auth/token", handler.GenerateToken)

	// Register handlers
	registerHandlersWithOptionsForTenant(r, handler, api_service.GinServerOptions{
		BaseURL: "/api/v1",
		Middlewares: []api_service.MiddlewareFunc{
			middleware.RequireAuth(jwt),
			middleware.RequireRole(auth.RoleAdmin),
		},
	})

	// Start server
	s := &http.Server{
		Addr:    cfg.GetURLBase(),
		Handler: r,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", err)
		return 1
	}

	select {
	case <-ctx.Done():
		logger.Info("timeout of 5 seconds")
	}
	logger.Info("Server Exited")
	return 0
}

func main() {
	start()
}
