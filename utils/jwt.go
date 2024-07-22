package utils

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(email, username string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Problem loading configs...")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"username": username,
		"exp":    time.Now().Add(time.Hour * 2).Unix(),
	})
	return token.SignedString([]byte(config.Secret))
}

func VerifyToken(token string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Problem loading configs...")
	}
	parsedToken, err := jwt.Parse(token, func(tkn *jwt.Token) (any, error) {
		_, isValidSigningMethod := tkn.Method.(*jwt.SigningMethodHMAC)
		if !isValidSigningMethod {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.Secret), nil
	})
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
