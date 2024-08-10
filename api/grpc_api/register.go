package grpc_api

import (
	"context"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	hashedPassword, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	args := db.CreateUserParams{
		Username: req.GetUsername(),
		Password: hashedPassword,
		Fullname: req.GetFullName(),
		Email:    req.GetEmail(),
	}
	user, err := server.store.CreateUser(ctx, args)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot register user: %s", err)
	}
	return &pb.RegisterUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.Fullname,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(user.CreatedAt.Time),
		},
	}, nil
}
