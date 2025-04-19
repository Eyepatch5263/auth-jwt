package routes

import (
	controller "github.com/eyepatch5263/auth_jwt/controllers"
	"github.com/gin-gonic/gin"
	"github.com/eyepatch5263/auth_jwt/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middlewares.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}