package rest_api

import (
	"database/sql"
	"net/http"
	"time"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/utils"
	"github.com/absk07/Go-Bank/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
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

	args := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username: req.Username,
			Password: hashedPassword,
			Fullname: req.FullName,
			Email:    req.Email,
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCrirical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"username":            txResult.User.Username,
			"fullname":            txResult.User.Fullname,
			"email":               txResult.User.Email,
			"password_changed_at": txResult.User.PasswordChangedAt,
			"created_at":          txResult.User.CreatedAt,
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
	var refreshTokenId uuid.UUID
	var token_expiration, refreshToken_expiration pgtype.Timestamptz
	_, token, token_expiration, err = utils.GenerateToken(
		user.Username,
		server.config.TokenDuration,
		server.config.Secret,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	refreshTokenId, refreshToken, refreshToken_expiration, err = utils.GenerateToken(
		user.Username,
		server.config.RefereshTokenDuration,
		server.config.Secret,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	// fmt.Printf("refreshToken_expiration: %+v\n", refreshToken_expiration)
	SessionParams := db.CreateSessionParams{
		ID:           refreshTokenId,
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

type verifyEmailReq struct {
	EmailId int64 `form:"email_id" binding:"required"`
	SecretCode string `form:"secret_code" binding:"required"`
}

func (server *Server) VerifyEmail(ctx *gin.Context) {
	var req verifyEmailReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.EmailId,
		SecretCode: req.SecretCode,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"isVerified": txResult.User.IsEmailVerified,
	})
}