package main

import (
	"os"
	routes "github.com/eyepatch5263/auth_jwt/routes"
	"github.com/gin-gonic/gin"
)

func main(){
	port:=os.Getenv("PORT")
	if port==""{
		port="8000"
	}

	router:=gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1",func(c *gin.Context){
		c.JSON(200,gin.H{
			"success":"Access ranted to api-1",
		})
	},)
	router.GET("/api-2",func(c *gin.Context){
		c.JSON(200,gin.H{
			"success":"Access ranted to api-2",
		})
	},)

	router.Run(":" + port)
}