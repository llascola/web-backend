package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/llascola/web-backend/internal/config"
)

// SecurityMiddleware checks the BearerAuthScopes value set by the generated
// OpenAPI wrapper. If scopes are present, it validates the JWT and enforces
// role-based access control. Public endpoints (no scopes in context) pass
// through without authentication.
func SecurityMiddleware(keys map[string]config.JWTKey, blocklist outports.TokenBlocklist) openapi.MiddlewareFunc {
	return func(c *gin.Context) {
		// If the generated wrapper did not set BearerAuthScopes,
		// this is a public endpoint — let it through.
		scopesRaw, exists := c.Get(openapi.BearerAuthScopes)
		if !exists {
			return
		}

		requiredScopes, _ := scopesRaw.([]string)

		// ── JWT Extraction ──────────────────────────────────

		var tokenString string
		var ok bool
		authHeader := c.GetHeader("Authorization")

		tokenString, ok = strings.CutPrefix(authHeader, "Bearer ")

		if !ok {
			cookieToke, err := c.Cookie("token")
			if err == nil && cookieToke != "" {
				tokenString = cookieToke
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// ── JWT Validation ──────────────────────────────────
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
		}, jwt.WithIssuer("portfolio-api"), jwt.WithAudience("portfolio-frontend"))

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Set claims to context so handlers can use them
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Check if token is blocked (revoked)
		if jti, ok := claims["jti"].(string); ok {
			blocked, err := blocklist.IsBlocked(c.Request.Context(), jti)
			if err != nil {
				// Fail open or closed? Security best practice: fail closed
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token status"})
				return
			}
			if blocked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				return
			}
		} else {
			// Reject tokens without a jti
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing jti"})
			return
		}

		c.Set("userID", claims["sub"])
		c.Set("role", claims["role"])
		c.Set("jti", claims["jti"])
		c.Set("exp", claims["exp"])

		// ── RBAC: Check required scopes ─────────────────────
		// If the endpoint requires specific scopes (e.g. "admin"),
		// verify the user's role matches.
		if len(requiredScopes) > 0 {
			userRole, _ := claims["role"].(string)
			authorized := slices.Contains(requiredScopes, userRole)

			if !authorized {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
				return
			}
		}
	}
}
