package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(dat), nil
}

func ValidatePassword(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
	return err
}

func IssueJWT(userId int, signingSecret string, expiresIn time.Duration) (string, error) {
	issuedTime := jwt.NewNumericDate(time.Now().UTC())
	expiredTime := jwt.NewNumericDate(issuedTime.Add(expiresIn * time.Second))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  issuedTime,
			ExpiresAt: expiredTime,
			Subject:   fmt.Sprintf("%v", userId),
		})

	signedToken, err := token.SignedString([]byte(signingSecret))

	if err != nil {
		return "", errors.New("unable to sign token")
	}

	return signedToken, nil
}

func ValidateJWT(token string, signingSecret string) (jwt.Token, jwt.Claims, error) {
	// parse the token, ensuring it's valid
	// claims can be parsed directly - type hinting included or off the token (no type knowledge)
	claims := jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(signingSecret), nil
		})
	if err != nil {
		return jwt.Token{}, nil, err
	}

	return *parsedToken, claims, nil

}

func ParseBearerToken(headers http.Header) (string, error) {
	// get token from request
	bearerToken := headers.Get("Authorization")

	if bearerToken == "" {
		return "", errors.New("no bearer token provided")
	}
	token := strings.TrimPrefix(bearerToken, "Bearer ")

	return token, nil
}
