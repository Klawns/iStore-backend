package main

import "github.com/gin-gonic/gin"

func registerRoutes(router *gin.Engine, dependencies dependencies) {
	router.POST("/users", dependencies.userHandler.Create)
	router.GET("/users/me", dependencies.authMiddleware.Authenticate(), dependencies.userHandler.Me)

	router.POST("/auth/sign-in", dependencies.authHandler.SignIn)
	router.POST("/auth/sign-out", dependencies.authHandler.SignOut)

	customers := router.Group("/customers", dependencies.authMiddleware.Authenticate())
	customers.POST("", dependencies.customerHandler.Create)
	customers.GET("", dependencies.customerHandler.List)
	customers.GET("/:id", dependencies.customerHandler.GetByID)
	customers.PUT("/:id", dependencies.customerHandler.Update)
	customers.DELETE("/:id", dependencies.customerHandler.Delete)

	sales := router.Group("/sales", dependencies.authMiddleware.Authenticate())
	sales.POST("", dependencies.saleHandler.Create)
	sales.GET("", dependencies.saleHandler.List)
	sales.GET("/installments/alerts", dependencies.saleHandler.ListInstallmentAlerts)
	sales.PATCH("/installments/:id/status", dependencies.saleHandler.UpdateInstallmentStatus)
	sales.GET("/period", dependencies.saleHandler.ListByPeriod)
	sales.GET("/:id/installments", dependencies.saleHandler.ListInstallmentsBySaleID)
	sales.GET("/:id", dependencies.saleHandler.GetByID)
	sales.PATCH("/:id/status", dependencies.saleHandler.UpdateStatus)
	sales.DELETE("/:id", dependencies.saleHandler.Delete)

	analytics := router.Group("/analytics", dependencies.authMiddleware.Authenticate())
	analytics.GET("/dashboard", dependencies.analyticsHandler.Dashboard)
	analytics.GET("/revenue", dependencies.analyticsHandler.Revenue)
	analytics.GET("/profit", dependencies.analyticsHandler.Profit)
	analytics.GET("/products/top", dependencies.analyticsHandler.TopProducts)
	analytics.GET("/payments", dependencies.analyticsHandler.PaymentMethods)
	analytics.GET("/customers/top", dependencies.analyticsHandler.TopCustomers)
	analytics.GET("/statuses", dependencies.analyticsHandler.Statuses)
}
