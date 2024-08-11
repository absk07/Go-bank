package grpc_api

import (
	"context"
	"database/sql"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "internal server error: %s", err)
	}
	IsPasswordValid := utils.IsPasswordValid(req.GetPassword(), user.Password)
	if !IsPasswordValid {
		return nil, status.Errorf(codes.NotFound, "wrong credentials: %s", err)
	}
	var token, refreshToken string
	var refreshTokenId uuid.UUID
	var token_expiration, refreshToken_expiration pgtype.Timestamptz
	_, token, token_expiration, err = utils.GenerateToken(
		user.Username,
		server.config.TokenDuration,
		server.config.Secret,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create token: %s", err)
	}
	refreshTokenId, refreshToken, refreshToken_expiration, err = utils.GenerateToken(
		user.Username,
		server.config.RefereshTokenDuration,
		server.config.Secret,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %s", err)

	}
	// fmt.Printf("refreshToken_expiration: %+v\n", refreshToken_expiration)
	SessionParams := db.CreateSessionParams{
		ID:           refreshTokenId,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
		ExpiresAt:    refreshToken_expiration,
	}
	session, err := server.store.CreateSession(ctx, SessionParams)
	// fmt.Printf("sessionParams: %+v\n", SessionParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err)
	}
	return &pb.LoginUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.Fullname,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(user.CreatedAt.Time),
		},
		SessionId:             session.ID.String(),
		AccessToken:           token,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(token_expiration.Time),
		RefreshTokenExpiresAt: timestamppb.New(refreshToken_expiration.Time),
	}, nil
}
