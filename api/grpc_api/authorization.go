package grpc_api

import (
	"context"
	"fmt"
	"strings"

	"github.com/absk07/Go-Bank/utils"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}
	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}
	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return "", fmt.Errorf("unsupported authorization type: %s", authType)
	}
	accessToken := fields[1]
	_, payload, err := utils.VerifyToken(accessToken, server.config.Secret)
	if err != nil {
		return "", fmt.Errorf("invalid access token: %s", err)
	}
	return payload, nil
}
