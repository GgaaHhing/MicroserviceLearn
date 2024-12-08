package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testProject/microservice/jwt_op"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" || len(token) == 0 {
			// 未经授权 401
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "认证失败，需要登陆",
			})
			c.Abort()
			return
		}
		j := jwt_op.NewJwt()
		tokenStr, err := j.ParseToken(token)
		if err != nil {
			if err.Error() == jwt_op.TokenExpired {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": jwt_op.TokenExpired,
				})
				c.Abort()
				return
				//refreshToken, err := j.RefreshToken(token)
				//if err != nil {
				//	return
				//}
				//c.Set("token", refreshToken)
				//c.Next()
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("token", tokenStr)
		c.Next()
	}
}
