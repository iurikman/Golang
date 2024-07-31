package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iurikman/smartSurvey/internal/models"
	log "github.com/sirupsen/logrus"
)

const (
	headerLength int = 2
)

func (s *Server) JWTAuth(next http.Handler) http.Handler {
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		claims, err := getClaimsFromHeader(r.Header.Get("Authorization"), s.key)

		switch {
		case errors.Is(err, models.ErrInvalidAccessToken):
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")

			return
		case err != nil:
			log.Warn("getClaimsFromHeader(r.Header.Get(\"Authorization\"), s.key) err")
			writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

			return
		}

		expiresAtTime := time.Unix(claims.ExpiresAt.Unix(), 0)
		if expiresAtTime.Before(time.Now()) {
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")

			return
		}

		userInfo := models.UserInfo{
			ID: claims.UUID,
		}

		r = r.WithContext(context.WithValue(r.Context(), models.UserInfoKey, userInfo))
		next.ServeHTTP(w, r)
	}

	return fn
}

func getClaimsFromHeader(authHeader string, key *rsa.PublicKey) (*models.Claims, error) {
	if authHeader == "" {
		return nil, models.ErrHeaderIsEmpty
	}

	headerParts := strings.Split(authHeader, " ")

	switch {
	case len(headerParts) != headerLength:
		return nil, models.ErrInvalidAccessToken
	case headerParts[0] != "Bearer":
		return nil, models.ErrInvalidAccessToken
	}

	claims, err := parseToken(headerParts[1], key)
	if err != nil {
		return nil, fmt.Errorf("parseToken(headerParts[1], key) err: %w", err)
	}

	return claims, nil
}

func parseToken(accessToken string, key *rsa.PublicKey) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, models.ErrInvalidAccessToken
		}

		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt.ParseWithClaims err: %w", err)
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok && token.Valid {
		return nil, models.ErrInvalidAccessToken
	}

	return claims, nil
}
