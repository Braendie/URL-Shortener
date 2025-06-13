package jwt

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Braendie/url-shortener/internal/clients/sso/grpc"
	"github.com/Braendie/url-shortener/internal/config"
	"github.com/Braendie/url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/net/context"
)

type email string
type userID string

func New(cfg *config.Config, log *slog.Logger, ssoClient *grpc.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.auth.jwt.New"

			log = log.With(
				slog.String("op", op),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Info("unauthorized request: missing token")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Info("unauthorized request: invalid signing method", slog.String("method", token.Header["alg"].(string)))
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.AppSecret), nil
			})
			if err != nil {
				log.Info("failed to parse token", sl.Err(err))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !token.Valid {
				log.Info("unauthorized request: invalid token")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Info("unauthorized request: invalid claims")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			isAdmin, err := ssoClient.IsAdmin(r.Context(), int64(claims["uid"].(float64)))
			if err != nil {
				log.Info("failed to check isAdmin")
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			if !isAdmin {
				log.Info("Not admin")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			log.Info("User is admin", slog.Int64("uid", int64(claims["uid"].(float64))))

			ctx := context.WithValue(r.Context(), email("email"), claims["email"].(string))
			ctx = context.WithValue(ctx, userID("user_id"), int64(claims["uid"].(float64)))
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
