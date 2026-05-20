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
}
