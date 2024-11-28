package middleware

import "github.com/gin-gonic/gin"

func responseWithError(c *gin.Context, code int, message interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"status_code":    -1,
		"status_message": message,
	})
}
