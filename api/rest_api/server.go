package rest_api

import (
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/middlewares"
	"github.com/absk07/Go-Bank/utils"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config utils.Config
	store *db.Store
	router *gin.Engine
}

func NewServer(config utils.Config, store *db.Store) *Server {
	server := &Server{
		config: config,
		store: store,
	}
	router := gin.Default()

	// routes
	router.POST("/v1/register", server.register)
	router.POST("/v1/login", server.login)
	router.POST("/v1/tokens/renew_access", server.renewAccessToken)
	router.POST("/v1/account/add", middlewares.IsAuthenticated, server.createAccount)
	router.GET("/v1/accounts", middlewares.IsAuthenticated, server.getAccounts)
	router.GET("/v1/account/:id", middlewares.IsAuthenticated, server.getAccountById)
	router.POST("/v1/transfer/add", middlewares.IsAuthenticated, server.createTransfer)

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}