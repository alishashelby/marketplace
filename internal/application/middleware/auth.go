package middleware

import (
	"context"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/pkg"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

const (
	reportMissingAuthorizationHeader = "no authorization header in request"
	reportParsingError               = "failed to parse bearer token"
	reportMissingUserKey             = "missing user key"
	reportInvalidUserData            = "invalid user data"
	reportMissingUserIDKey           = "no appropriate ID key found in bearer token"
	reportUnexpectedStringError      = "unexpected format of userID"
	reportInvalidUserIDKey           = "userID should be a UUID"
)

func AuthMiddleware(jwtService *service.JWTService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("AuthMiddleware")

		tokenString := strings.TrimSpace(
			strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "),
		)

		if tokenString == "" {
			pkg.SendJSON(w, http.StatusUnauthorized, reportMissingAuthorizationHeader)
			return
		}

		jwtClaims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			pkg.SendJSON(w, http.StatusUnauthorized, reportParsingError)
			return
		}

		userClaim, exists := jwtClaims[string(service.UserKey)]
		if !exists {
			pkg.SendJSON(w, http.StatusUnauthorized, reportMissingUserKey)
			return
		}

		userMap, ok := userClaim.(map[string]interface{})
		if !ok {
			pkg.SendJSON(w, http.StatusUnauthorized, reportInvalidUserData)
			return
		}

		rawUserID, exists := userMap[string(service.UserIDKey)]
		if !exists {
			pkg.SendJSON(w, http.StatusUnauthorized, reportMissingUserIDKey)
			return
		}

		stringUserID, ok := rawUserID.(string)
		if !ok {
			pkg.SendJSON(w, http.StatusUnauthorized, reportUnexpectedStringError)
			return
		}

		userID, err := uuid.Parse(stringUserID)
		if err != nil {
			pkg.SendJSON(w, http.StatusUnauthorized, reportInvalidUserIDKey)
			return
		}

		ctx := context.WithValue(r.Context(), service.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
