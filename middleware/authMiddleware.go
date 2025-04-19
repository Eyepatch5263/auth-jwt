package middlewares

import ("github.com/gin-gonic/gin"
	"net/http"
	helper "github.com/eyepatch5263/auth_jwt/helpers"

)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token:=c.Request.Header.Get("token")
		if token==""{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"no authorization header provided"})
			c.Abort()
			return
		}
		claims,err:=helper.ValidateToken(token)
		if err!=""{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"Unauthorized access"})
			c.Abort()
			return
		}
		c.Set("email",claims.Email)
		c.Set("first_name",claims.First_name)
		c.Set("last_name",claims.Last_name)
		c.Set("uid",claims.Uid)
		c.Set("user_type",claims.User_type)
		c.Next()
	}
}