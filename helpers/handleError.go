package helpers

import "github.com/gin-gonic/gin"

func ErrorResponse(err error) gin.H {
	return gin.H{
		"error": true,
		"message": err.Error(),
	}
}