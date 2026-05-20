package main

import (
	authHandler "istore/internal/auth/handler"
	authMiddleware "istore/internal/auth/middleware"
	authImplementation "istore/internal/auth/service/implementation"
	"istore/internal/users/entity"
	userHandler "istore/internal/users/handler"
	userRepositoryImplementation "istore/internal/users/repository/implementation"
	userServiceImplementation "istore/internal/users/service/implementation"
	sqliteDB "istore/pkg/database/sqlite"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := getEnv("DB_PATH", "istore.db")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-me")
	port := getEnv("PORT", "8080")

	db, err := sqliteDB.Connect(dbPath)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err := db.AutoMigrate(&entity.UserEntity{}); err != nil {
		log.Fatalf("error running migrations: %v", err)
	}

	jwtProvider := authImplementation.NewJwtService(jwtSecret)
	cookieManager := authImplementation.NewCookieService()
	userRepository := userRepositoryImplementation.NewUserRepository(db)
	userService := userServiceImplementation.NewUserService(userRepository)
	authService := authImplementation.NewAuthService(userRepository, jwtProvider)
	authMw := authMiddleware.NewAuthMiddleware(jwtProvider, cookieManager)

	usersHandler := userHandler.NewUserHandler(userService)
	authHandlerInstance := authHandler.NewAuthHandler(authService, cookieManager)

	router := gin.Default()
	router.POST("/users", usersHandler.Create)
	router.POST("/auth/sign-in", authHandlerInstance.SignIn)
	router.POST("/auth/sign-out", authHandlerInstance.SignOut)
	router.GET("/users/me", authMw.Authenticate(), usersHandler.Me)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
