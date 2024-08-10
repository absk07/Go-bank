package rest_api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/gin-gonic/gin"
)

type transferReq struct {
	FromAccountId int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountId   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return
	}
	fromAccount, valid := server.validAccount(ctx, req.FromAccountId, req.Currency)
	if !valid {
		return
	}
	current_user := ctx.GetString("username")
	if fromAccount.Owner != current_user {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Account does'nt belong to authenticated user!",
		})
		return
	}
	_, valid = server.validAccount(ctx, req.ToAccountId, req.Currency)
	if !valid {
		return
	}
	args := db.TransferTxParams{
		FromAccountId: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        req.Amount,
	}
	res, err := server.store.TransferTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, helpers.ErrorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, helpers.ErrorResponse(err))
		return account, false
	}
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, helpers.ErrorResponse(err))
		return account, false
	}
	return account, true
}
