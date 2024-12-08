package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CrossDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		//允许的来源
		c.Header("Access-Control-Allow-Origin",
			"*")
		//允许的方法
		c.Header("Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, DELETE")
		//允许过来的请求携带的请求头
		c.Header("Access-Control-Allow-Headers",
			"Origin, X-Requested-With, Content-Type, AccessToken, Authorization, Token")
		//
		c.Header("Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin,"+
				"Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials",
			"true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}
