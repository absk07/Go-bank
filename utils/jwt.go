package utils

import (
	"errors"
	// "fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GenerateToken(username string, duration time.Duration, secret string) (uuid.UUID, string, pgtype.Timestamptz, error) {
	tokenId := uuid.New()
	expirationTime := pgtype.Timestamptz{
		Time:  time.Now().Add(duration),
		Valid: true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":       tokenId,
		"username": username,
		"exp":      expirationTime.Time.Unix(),
	})
	accessToken, err := token.SignedString([]byte(secret))
	return tokenId, accessToken, expirationTime, err
}

func VerifyToken(token, secret string) (uuid.UUID, string, error) {
	parsedToken, err := jwt.Parse(token, func(tkn *jwt.Token) (any, error) {
		_, isValidSigningMethod := tkn.Method.(*jwt.SigningMethodHMAC)
		if !isValidSigningMethod {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	// fmt.Println("parsedToken", parsedToken)
	if err != nil {
		return uuid.UUID{}, "", errors.New("could not parse token")
	}
	isTokenValid := parsedToken.Valid
	if !isTokenValid {
		return uuid.UUID{}, "", errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.UUID{}, "", errors.New("invalid token claims")
	}
	// fmt.Println("Claims", claims)
	id := claims["ID"].(string)
	username := claims["username"].(string)
	return uuid.MustParse(id), username, nil
}
