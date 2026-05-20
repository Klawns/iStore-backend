package main

import (
	authHandler "istore/internal/auth/handler"
	authMiddleware "istore/internal/auth/middleware"
	authImplementation "istore/internal/auth/service/implementation"
	customerHandler "istore/internal/customer/handler"
	customerRepositoryImplementation "istore/internal/customer/repository/implementation"
	customerServiceImplementation "istore/internal/customer/service/implementation"
	userHandler "istore/internal/users/handler"
	userRepositoryImplementation "istore/internal/users/repository/implementation"
	userServiceImplementation "istore/internal/users/service/implementation"

	"gorm.io/gorm"
)

type dependencies struct {
	authHandler     *authHandler.AuthHandler
	authMiddleware  *authMiddleware.AuthMiddleware
	customerHandler *customerHandler.CustomerHandler
	userHandler     *userHandler.UserHandler
}

func buildDependencies(db *gorm.DB) dependencies {
	jwtProvider := authImplementation.NewJwtService(getEnv("JWT_SECRET", "dev-secret-change-me"))
	cookieManager := authImplementation.NewCookieService()

	userRepository := userRepositoryImplementation.NewUserRepository(db)
	userService := userServiceImplementation.NewUserService(userRepository)
	authService := authImplementation.NewAuthService(userRepository, jwtProvider)

	customerRepository := customerRepositoryImplementation.NewCustomerRepository(db)
	customerService := customerServiceImplementation.NewCustomerService(customerRepository)

	return dependencies{
		authHandler:     authHandler.NewAuthHandler(authService, cookieManager),
		authMiddleware:  authMiddleware.NewAuthMiddleware(jwtProvider, cookieManager),
		customerHandler: customerHandler.NewCustomerHandler(customerService),
		userHandler:     userHandler.NewUserHandler(userService),
	}
}
