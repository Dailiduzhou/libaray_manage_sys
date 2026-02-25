package routes

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, userHandler *controller.UserHandler) {
	api := r.Group("/api")
	{

		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		authGroup := api.Group("/")
		authGroup.Use(middleware.AuthRequired())
		{
			authGroup.POST("/logout", userHandler.Logout)
		}
	}
}
