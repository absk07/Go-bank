package grpc_api

import (
	"context"
	"time"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/absk07/Go-Bank/worker"
	"github.com/hibiken/asynq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, helpers.InvalidArgumentError(violations)
	}

	hashedPassword, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	args := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username: req.GetUsername(),
			Password: hashedPassword,
			Fullname: req.GetFullName(),
			Email:    req.GetEmail(),
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
		return nil, status.Errorf(codes.Internal, "cannot register user: %s", err)
	}

	return &pb.RegisterUserResponse{
		User: &pb.User{
			Username:          txResult.User.Username,
			FullName:          txResult.User.Fullname,
			Email:             txResult.User.Email,
			PasswordChangedAt: timestamppb.New(txResult.User.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(txResult.User.CreatedAt.Time),
		},
	}, nil
}

func validateCreateUserRequest(req *pb.RegisterUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := utils.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, helpers.FieldViolation("username", err))
	}
	if err := utils.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, helpers.FieldViolation("password", err))
	}
	if err := utils.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, helpers.FieldViolation("full_name", err))
	}
	if err := utils.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, helpers.FieldViolation("email", err))
	}
	return violations
}