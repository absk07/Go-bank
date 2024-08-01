package utils

import (
	"errors"
	// "fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GenerateToken(email string, username string, duration time.Duration, secret string) (string, pgtype.Timestamptz, error) {
	expirationTime := pgtype.Timestamptz{
		Time:  time.Now().Add(duration),
		Valid: true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":    email,
		"username": username,
		"exp":      expirationTime.Time.Unix(),
	})
	accessToken, err := token.SignedString([]byte(secret))
	return accessToken, expirationTime, err
}

func VerifyToken(token, secret string) (string, error) {
	parsedToken, err := jwt.Parse(token, func(tkn *jwt.Token) (any, error) {
		_, isValidSigningMethod := tkn.Method.(*jwt.SigningMethodHMAC)
		if !isValidSigningMethod {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	// fmt.Println("parsedToken", parsedToken)
	if err != nil {
		return "", errors.New("could not parse token")
	}
	isTokenValid := parsedToken.Valid
	if !isTokenValid {
		return "", errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	// fmt.Println("Claims", claims)
	// email := claims["email"].(string)
	username := claims["username"].(string)
	return username, nil
}
