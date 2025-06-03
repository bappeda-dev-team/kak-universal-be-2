package middleware

import (
	"context"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/web"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/MicahParks/keyfunc"
)

var JWKS *keyfunc.JWKS

func InitJWKS(jwksURL string) error {
	var err error
	JWKS, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
	})
	return err
}

type AuthMiddleware struct {
	Handler http.Handler
}

func NewAuthMiddleware(handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{Handler: handler}
}


func (middleware *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" || !strings.HasPrefix(tokenHeader, "Bearer ") {
		writeUnauthorized(w, "Missing or invalid Authorization header")
		return
	}
	rawToken := strings.TrimPrefix(tokenHeader, "Bearer ")


	token, err := jwt.Parse(rawToken, JWKS.Keyfunc)
	if err != nil || !token.Valid {
		writeUnauthorized(w, "Invalid token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeUnauthorized(w, "Invalid claims")
		return
	}

	// Optional: validate audience or issuer
	issuer := os.Getenv("KEYCLOAK_ISSUER")
	if iss, ok := claims["iss"].(string); !ok || iss != issuer {
		writeUnauthorized(w, "Invalid issuer")
		return
	}

	// Set user info in context
	ctx := context.WithValue(r.Context(), helper.UserInfoKey, claims)
	r = r.WithContext(ctx)

	middleware.Handler.ServeHTTP(w, r)
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	helper.WriteToResponseBody(w, web.WebResponse{
		Code:   http.StatusUnauthorized,
		Status: "UNAUTHORIZED",
		Data:   message,
	})
}
