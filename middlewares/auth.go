package middlewares

import (
	// "fmt"
	"log"
	"net/http"
	"strings"

	"github.com/absk07/Go-Bank/utils"
	"github.com/gin-gonic/gin"
)

func IsAuthenticated(ctx *gin.Context) {
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatal("Problem loading configs...")
	}
	auth_header := ctx.Request.Header.Get("Authorization")
	fields := strings.Fields(auth_header)
	if len(fields) < 2 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authorization header format required!",
		})
		return
	}
	if strings.ToLower(fields[0]) != strings.ToLower("Bearer") {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid authorization header format!",
		})
		return
	}
	token := fields[1]
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "1 User Unauthenticated!",
		})
		return
	}
	username, err := utils.VerifyToken(token, config.Secret)
	// fmt.Println(err)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "2 User Unauthenticated!",
		})
		return
	}
	ctx.Set("username", username)
	ctx.Next()
}