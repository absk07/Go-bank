package api

import (
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/middlewares"
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
	router.POST("/users/register", server.register)
	router.POST("/users/login", server.login)
	router.POST("/account/add", middlewares.IsAuthenticated, server.createAccount)
	router.GET("/accounts", middlewares.IsAuthenticated, server.getAccounts)
	router.GET("/account/:id", middlewares.IsAuthenticated, server.getAccountById)
	router.POST("/transfer/add", middlewares.IsAuthenticated, server.createTransfer)

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}