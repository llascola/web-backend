package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/config"
)

func AuthMiddleware(keys map[string]config.JWTKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing key id (kid) in token header")
			}

			keyConfig, exists := keys[kid]
			if !exists {
				return nil, fmt.Errorf("unknown key id: %v", kid)
			}

			if token.Method.Alg() != keyConfig.Algorithm {
				return nil, fmt.Errorf("unexpected signing method: %v, expected: %v", token.Method.Alg(), keyConfig.Algorithm)
			}

			return keyConfig.Secret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Set claims to context so controllers can use it
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("userID", claims["sub"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}

func RequireRole(requiredRole domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Check if the user's role matches the required role
		if domain.UserRole(userRole.(string)) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
			return
		}

		c.Next()
	}
}
