package api

import (
	"database/sql"
	"net/http"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserReq struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=3"`
	FullName string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (server *Server) register(ctx *gin.Context) {
	var req createUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}

	args := db.CreateUserParams{
		Username: req.Username,
		Password: hashedPassword,
		Fullname: req.FullName,
		Email:    req.Email,
	}
	user, err := server.store.CreateUser(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"username":            user.Username,
			"fullname":            user.Fullname,
			"email":               user.Email,
			"password_changed_at": user.PasswordChangedAt,
			"created_at":          user.CreatedAt,
		},
	})
}

type loginUserReq struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=3"`
}

func (server *Server) login(ctx *gin.Context) {
	var req loginUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, helpers.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	IsPasswordValid := utils.IsPasswordValid(req.Password, user.Password)
	if !IsPasswordValid {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "invalid credentials",
		})
		return
	}
	var token, refreshToken string
	var token_expiration, refreshToken_expiration pgtype.Timestamptz
	token, token_expiration, err = utils.GenerateToken(user.Email, user.Username, server.config.TokenDuration, server.config.Secret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	refreshToken, refreshToken_expiration, err = utils.GenerateToken(user.Email, user.Username, server.config.RefereshTokenDuration, server.config.Secret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	// fmt.Printf("refreshToken_expiration: %+v\n", refreshToken_expiration)	
	SessionParams := db.CreateSessionParams{
		ID:           uuid.New(),
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshToken_expiration,
	}
	session, err := server.store.CreateSession(ctx, SessionParams)
	// fmt.Printf("sessionParams: %+v\n", SessionParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessionID":                session.ID,
			"username":                 user.Username,
			"fullname":                 user.Fullname,
			"email":                    user.Email,
			"password_changed_at":      user.PasswordChangedAt,
			"created_at":               user.CreatedAt,
			"token":                    token,
			"token_expires_at":         token_expiration.Time,
			"refresh_token":            refreshToken,
			"refresh_token_expires_at": refreshToken_expiration.Time,
		},
	})
}
