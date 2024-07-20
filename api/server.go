package api

import (
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// routes
	router.POST("/users/register", server.createUser)
	router.POST("/account/add", server.createAccount)
	router.GET("/accounts", server.getAccounts)
	router.GET("/account/:id", server.getAccountById)
	router.POST("/transfer/add", server.createTransfer)

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}