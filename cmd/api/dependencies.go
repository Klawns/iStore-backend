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
	privacyHandler "istore/internal/privacy/handler"
	privacyRepositoryImplementation "istore/internal/privacy/repository/implementation"
	privacyServiceImplementation "istore/internal/privacy/service/implementation"
	saleHandler "istore/internal/sale/handler"
	saleRepositoryImplementation "istore/internal/sale/repository/implementation"
	saleServiceContract "istore/internal/sale/service/contract"
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
	privacyHandler   *privacyHandler.PrivacyHandler
	saleHandler      *saleHandler.SaleHandler
	saleService      saleServiceContract.SaleService
	userHandler      *userHandler.UserHandler
}

func buildDependencies(db *gorm.DB, jwtSecret string) dependencies {
	jwtProvider := authImplementation.NewJwtService(jwtSecret)
	cookieManager := authImplementation.NewCookieService(isProduction())

	userRepository := userRepositoryImplementation.NewUserRepository(db)
	userService := userServiceImplementation.NewUserService(userRepository)
	authService := authImplementation.NewAuthService(userRepository, jwtProvider)

	customerRepository := customerRepositoryImplementation.NewCustomerRepository(db)
	customerService := customerServiceImplementation.NewCustomerService(customerRepository)

	privacyRepository := privacyRepositoryImplementation.NewPrivacyRepository(db)
	privacyService := privacyServiceImplementation.NewPrivacyService(privacyRepository)

	saleRepository := saleRepositoryImplementation.NewSaleRepository(db)
	saleService := saleServiceImplementation.NewSaleService(saleRepository)

	analyticsRepository := analyticsRepositoryImplementation.NewAnalyticsRepository(db)
	analyticsService := analyticsServiceImplementation.NewAnalyticsService(analyticsRepository)

	return dependencies{
		analyticsHandler: analyticsHandler.NewAnalyticsHandler(analyticsService),
		authHandler:      authHandler.NewAuthHandler(authService, cookieManager),
		authMiddleware:   authMiddleware.NewAuthMiddleware(jwtProvider, cookieManager),
		customerHandler:  customerHandler.NewCustomerHandler(customerService),
		privacyHandler:   privacyHandler.NewPrivacyHandler(privacyService),
		saleHandler:      saleHandler.NewSaleHandler(saleService),
		saleService:      saleService,
		userHandler:      userHandler.NewUserHandler(userService, cookieManager),
	}
}
