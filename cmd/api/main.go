package main

import (
	postgresDB "istore/pkg/database/postgres"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file")
	}

	db, err := postgresDB.Connect(getDatabaseDSN())
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("error running migrations: %v", err)
	}

	dependencies := buildDependencies(db)

	router := gin.Default()
	registerRoutes(router, dependencies)

	if err := router.Run(":" + getEnv("PORT", "8080")); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
