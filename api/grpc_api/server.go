package grpc_api

import (
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
)

type Server struct {
	pb.UnimplementedGoBankServer 
	config utils.Config
	store *db.Store
}

func NewServer(config utils.Config, store *db.Store) *Server {
	server := &Server{
		config: config,
		store: store,
	}
	return server
}
