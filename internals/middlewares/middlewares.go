package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"petplate/utils"
)

// AuthMiddleware checks the token validity and extracts the claims
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Check if the token has the "Bearer " prefix and strip it if present
		if strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		}

		tokenString := strings.TrimSpace(authHeader)
		claims := &utils.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return utils.JWTKey, nil
		})

		if err != nil {
			fmt.Printf("Token parse error: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			c.Abort()
			return
		}

		// Set email in context
		c.Set("email", claims.Email)
		c.Next()
	}
}
