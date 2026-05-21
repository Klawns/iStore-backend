package main

import (
	analyticsHandler "istore/internal/analytics/handler"
	analyticsRepositoryImplementation "istore/internal/analytics/repository/implementation"
	analyticsServiceImplementation "istore/internal/analytics/service/implementation"
	authHandler "istore/internal/auth/handler"
	authMiddleware "istore/internal/auth/middleware"
	authImplementation "istore/internal/auth/service/implementation"
	customerHandler "istore/internal/customer/handler"
	customerRepositoryImplementation "istore/internal/customer/repository/implementation"
	customerServiceImplementation "istore/internal/customer/service/implementation"
	saleHandler "istore/internal/sale/handler"
	saleRepositoryImplementation "istore/internal/sale/repository/implementation"
	saleServiceImplementation "istore/internal/sale/service/implementation"
	userHandler "istore/internal/users/handler"
	userRepositoryImplementation "istore/internal/users/repository/implementation"
	userServiceImplementation "istore/internal/users/service/implementation"

	"gorm.io/gorm"
)

type dependencies struct {
	analyticsHandler *analyticsHandler.AnalyticsHandler
	authHandler      *authHandler.AuthHandler
	authMiddleware   *authMiddleware.AuthMiddleware
	customerHandler  *customerHandler.CustomerHandler
	saleHandler      *saleHandler.SaleHandler
	userHandler      *userHandler.UserHandler
}

func buildDependencies(db *gorm.DB) dependencies {
	jwtProvider := authImplementation.NewJwtService(getEnv("JWT_SECRET", "dev-secret-change-me"))
	cookieManager := authImplementation.NewCookieService()

	userRepository := userRepositoryImplementation.NewUserRepository(db)
	userService := userServiceImplementation.NewUserService(userRepository)
	authService := authImplementation.NewAuthService(userRepository, jwtProvider)

	customerRepository := customerRepositoryImplementation.NewCustomerRepository(db)
	customerService := customerServiceImplementation.NewCustomerService(customerRepository)

	saleRepository := saleRepositoryImplementation.NewSaleRepository(db)
	saleService := saleServiceImplementation.NewSaleService(saleRepository)

	analyticsRepository := analyticsRepositoryImplementation.NewAnalyticsRepository(db)
	analyticsService := analyticsServiceImplementation.NewAnalyticsService(analyticsRepository)

	return dependencies{
		analyticsHandler: analyticsHandler.NewAnalyticsHandler(analyticsService),
		authHandler:      authHandler.NewAuthHandler(authService, cookieManager),
		authMiddleware:   authMiddleware.NewAuthMiddleware(jwtProvider, cookieManager),
		customerHandler:  customerHandler.NewCustomerHandler(customerService),
		saleHandler:      saleHandler.NewSaleHandler(saleService),
		userHandler:      userHandler.NewUserHandler(userService),
	}
}
