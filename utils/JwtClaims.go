package utils

import (
	"errors"
	//"petplate/internals/models"
	//"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)
var JwtSecret = []byte("tokjwtsh")
func GetJWTClaim(c *gin.Context ) (email string, err error) {
    // Retrieve JWT from the "Authorization" cookie
    JWTToken, err := c.Cookie("Authorization")
    if JWTToken == "" || err != nil {
        return "", errors.New("no authorization token available")
    }

    // Define your JWT secret directly (instead of using GetEnvVariables)
    hmacSecret := []byte("njwtsh") // Replace with your actual secret key

    // Parse the token
    token, err := jwt.Parse(JWTToken, func(token *jwt.Token) (interface{}, error) {
        // Ensure the signing method is HMAC
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return hmacSecret, nil
    })

    if err != nil {
        return "", errors.New("request unauthorized")
    }

    // Extract and validate claims
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Check for expiration
        expirationTime, ok := claims["exp"].(float64)
        if !ok {
            return "", errors.New("request unauthorized")
        }

        expiration := time.Unix(int64(expirationTime), 0)
        if time.Now().After(expiration) {
            return "", errors.New("token has expired")
        }

        // Extract the email from claims
        email, ok := claims["email"].(string)
        if !ok {
            return "", errors.New("email not found in token claims")
        }
        return email, nil
    } else {
        return "", errors.New("invalid token claims")
    }
}
func GenerateJWT(email string) (string, error) {
    // Set the expiration time (1 day)
    expirationTime := time.Now().Add(24 * time.Hour)

    // Define the claims
    claims := jwt.MapClaims{
        "email": email,
        "exp":   expirationTime.Unix(),
    }

    // Create a new token with the claims
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign the token using a secret key (replace "njwtsh" with your actual secret)
    tokenString, err := token.SignedString([]byte("njwtsh")) // Replace with your actual secret key
    if err != nil {
        return "", err
    }

    return tokenString, nil
}




