package rest_api

import (
	"database/sql"
	"net/http"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/gin-gonic/gin"
)

type createAccountReq struct {
	// Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	current_user := ctx.GetString("username")
	args := db.CreateAccountParams{
		Owner:    current_user,
		Currency: req.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    account,
	})
}

type getAccountByIdReq struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccountById(ctx *gin.Context) {
	var req getAccountByIdReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, helpers.ErrorResponse(err))
		return
		}
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	current_user := ctx.GetString("username")
	if account.Owner != current_user{
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Account does'nt belong to authenticated user!",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    account,
	})
}


type getAccountsReq struct {
	Page int32 `form:"page" binding:"required,min=1"`
	Size int32 `form:"size" binding:"required,min=5,max=10"`
}

func (server *Server) getAccounts(ctx *gin.Context) {
	var req getAccountsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	current_user := ctx.GetString("username")
	args := db.ListAccountParams{
		Owner: current_user,
		Limit: req.Size,
		Offset: (req.Page - 1) * req.Size,
	}
	account, err := server.store.ListAccount(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    account,
	})
}