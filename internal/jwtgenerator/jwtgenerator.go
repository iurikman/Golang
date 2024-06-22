package jwtgenerator

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iurikman/smartSurvey/internal/models"
	log "github.com/sirupsen/logrus"
)

const (
	hoursInDay     = 24
	readerBits int = 4096
)

type JWTGenerator struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func NewJWTGenerator() *JWTGenerator {
	privateKey, err := rsa.GenerateKey(rand.Reader, readerBits)
	if err != nil {
		log.Warn("rsa.GenerateKey(rand.Reader, 4096)", err)
	}

	generator := &JWTGenerator{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	return generator
}

func (j *JWTGenerator) GetNewTokenString(user models.User) (string, error) {
	claims := models.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "user auth server",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * hoursInDay)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UUID: user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	ss, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("token.SignedString(j.privateKey) err: %w", err)
	}

	return ss, nil
}

func (j *JWTGenerator) GetPublicKey() *rsa.PublicKey {
	key := *j.publicKey

	return &key
}
