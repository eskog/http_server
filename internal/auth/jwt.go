package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String()})
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	} else if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.UUID{}, err
		}
		return id, nil
	}
	return uuid.UUID{}, errors.New("unable to verify JWT")
}

func MakeRefreshToken() (string, error) {
	var refreshToken [32]byte
	rand.Read(refreshToken[:])
	result := hex.EncodeToString(refreshToken[:])
	return result, nil

}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	token, ok := strings.CutPrefix(token, "Bearer ")
	if !ok {
		return "", errors.New("unable to extract token")
	}
	return token, nil

}
