package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/insurance-portal/poc/internal/domain/user"
	"github.com/insurance-portal/poc/internal/interfaces/dto"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	jwtSecret []byte
	userRepo  user.Repository
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtSecret string, userRepo user.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		userRepo:  userRepo,
	}
}

// Authenticate validates JWT token and adds user info to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Missing authorization header", nil)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Invalid authorization header format", nil)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid token", err)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			respondError(w, http.StatusUnauthorized, "Invalid token claims", nil)
			return
		}

		// Get user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			respondError(w, http.StatusUnauthorized, "Invalid user ID in token", nil)
			return
		}

		// Get user role from claims
		role, ok := claims["role"].(string)
		if !ok {
			respondError(w, http.StatusUnauthorized, "Invalid role in token", nil)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user_role", role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if the user has the required role
func (m *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("user_role").(string)
			if !ok {
				respondError(w, http.StatusUnauthorized, "User role not found in context", nil)
				return
			}

			if role != requiredRole {
				respondError(w, http.StatusForbidden, "Insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoles checks if the user has any of the required roles
func (m *AuthMiddleware) RequireRoles(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("user_role").(string)
			if !ok {
				respondError(w, http.StatusUnauthorized, "User role not found in context", nil)
				return
			}

			hasRole := false
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondError(w, http.StatusForbidden, "Insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}