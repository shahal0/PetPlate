package middlewares

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("tokjwtsh") // Replace with your secret key

// AuthMiddleware is the middleware function to verify JWT token
func AuthMiddleware() gin.HandlerFunc {
	log.Println("AuthMiddleware hit")
	return func(c *gin.Context) {
		// Extract the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		authHeader = strings.TrimSpace(authHeader) // Trim any spaces in the header
		log.Println("Received Authorization Header:", authHeader)
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// The token should be in the format "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Println("Extracted JWT Token:", tokenString)
		
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token is signed with the expected method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			log.Println("JWT Parsing Error:", err) // Log the error for debugging
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Validate token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Safely extract and validate the "exp" claim
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Expiration time missing in token"})
				c.Abort()
				return
			}

			// Set user information (e.g., email) in the context
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email in token"})
				c.Abort()
				return
			}

			// Add any additional claim processing here if needed (e.g., roles, user IDs)

		} else {
			log.Println("JWT Claims Invalid:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Proceed to the next handler if the token is valid
		c.Next()
	}
}

