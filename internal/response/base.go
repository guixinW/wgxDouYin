package response

import "github.com/gin-gonic/gin"

type StatusCode int

const (
	StatusTokenExpired StatusCode = iota
	StatusTokenInvalid
	StatusOther
)

func AbortWithError(c *gin.Context, code int, statusCode StatusCode, message interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"status_code":    statusCode,
		"status_message": message,
	})
}

type Base struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func ErrorResponse(errorMsg string) Base {
	return Base{
		StatusCode: -1,
		StatusMsg:  errorMsg,
	}
}
