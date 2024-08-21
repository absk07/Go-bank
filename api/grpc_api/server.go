package grpc_api

import (
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/absk07/Go-Bank/worker"
)

type Server struct {
	pb.UnimplementedGoBankServer 
	config utils.Config
	store *db.Store
	taskDistributor worker.TaskDistributor
}

func NewServer(config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) *Server {
	server := &Server{
		config: config,
		store: store,
		taskDistributor: taskDistributor,
	}
	return server
}
