package rest_api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/utils"
	"github.com/gin-gonic/gin"
)

type renewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	refresh_token_id, username, err := utils.VerifyToken(req.RefreshToken, server.config.Secret)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, helpers.ErrorResponse(err))
		return
	}
	session, err := server.store.GetSession(ctx, refresh_token_id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, helpers.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Session Blocked!",
		})
	}
	if session.Username != username {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Incorrect session user!",
		})
	}
	if session.RefreshToken != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Refresh token mismatch!",
		})
	}
	if time.Now().After(session.ExpiresAt.Time) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Expired session!",
		})
	}
	_, token, token_expiration, err := utils.GenerateToken(
		username,
		server.config.TokenDuration,
		server.config.Secret,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":            token,
			"token_expires_at": token_expiration.Time,
		},
	})
}
