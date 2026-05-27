package main

import (
	"context"
	"errors"
	postgresDB "istore/pkg/database/postgres"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const (
	databasePingTimeout = 5 * time.Second
	migrationTimeout    = 30 * time.Second
	readyzPingTimeout   = 3 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("error loading .env file")
	}

	log.Println("opening postgres connection")
	db, err := postgresDB.Connect(getDatabaseDSN())
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	log.Println("postgres connection opened")

	log.Println("pinging postgres")
	if err := pingDatabaseWithTimeout(db, databasePingTimeout); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}
	log.Println("postgres ping succeeded")

	log.Println("running database migrations")
	migrationCtx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()
	if err := runMigrationsWithContext(migrationCtx, db); err != nil {
		log.Fatalf("error running migrations: %v", err)
	}
	log.Println("database migrations completed")

	jwtSecret, err := getJWTSecret()
	if err != nil {
		log.Fatalf("error loading JWT secret: %v", err)
	}

	dependencies := buildDependencies(db, jwtSecret)

	router := gin.Default()
	registerRoutes(router, dependencies)
	registerDiagnosticsRoutes(router, db)

	port := getEnv("PORT", "8080")
	log.Printf("starting HTTP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

func registerDiagnosticsRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(ctx *gin.Context) {
		if err := pingDatabaseWithTimeout(db, readyzPingTimeout); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

func pingDatabaseWithTimeout(db *gorm.DB, timeout time.Duration) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return sqlDB.PingContext(ctx)
}
