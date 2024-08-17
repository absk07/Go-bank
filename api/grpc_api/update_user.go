package grpc_api

import (
	"context"
	"database/sql"
	"time"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/helpers"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	payload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, helpers.UnauthenticatedError(err)
	}

	if payload != req.Username {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info!")
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, helpers.InvalidArgumentError(violations)
	}

	args := db.UpdateUserParams{
		Username: req.GetUsername(),
		Fullname: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}

		args.Password = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		args.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}
	user, err := server.store.UpdateUser(ctx, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot register user: %s", err)
	}
	return &pb.UpdateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.Fullname,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(user.CreatedAt.Time),
		},
	}, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := utils.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, helpers.FieldViolation("username", err))
	}
	if req.Password != nil {
		if err := utils.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, helpers.FieldViolation("password", err))
		}
	}
	if req.FullName != nil {
		if err := utils.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, helpers.FieldViolation("full_name", err))
		}
	}
	if req.Email != nil {
		if err := utils.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, helpers.FieldViolation("email", err))
		}
	}
	return violations
}
