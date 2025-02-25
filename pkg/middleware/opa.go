package middleware

import (
	"github.com/gin-gonic/gin"
)

var OpaUrl = "http://localhost:8181/v1/data/authz/allow"

func OPAAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
