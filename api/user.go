package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/utils"
	"github.com/gin-gonic/gin"
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
			"username": user.Username,
			"fullname": user.Fullname,
			"email":    user.Email,
			"password_changed_at": user.PasswordChangedAt,
			"created_at": user.CreatedAt,
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
	fmt.Print("body-", req.Password, "user-", user.Password)
	if !IsPasswordValid {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": true,
			"message": "invalid credentials",
		})
		return
	}
	var token string
	token, err = utils.GenerateToken(user.Email, user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"username": user.Username,
			"fullname": user.Fullname,
			"email":    user.Email,
			"password_changed_at": user.PasswordChangedAt,
			"created_at": user.CreatedAt,
			"token": token,
		},
	})
}