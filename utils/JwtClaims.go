package utils

import (
	//"errors"
	"strings"
	//"petplate/internals/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)
var JwtSecret = []byte("tokjwtsh")

func GetJWTClaim(c *gin.Context) (*jwt.MapClaims, error) {
    authHeader := c.GetHeader("Authorization")
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, jwt.ErrSignatureInvalid
        }
        return JwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
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
    tokenString, err := token.SignedString([]byte("tokjwtsh")) // Replace with your actual secret key
    if err != nil {
        return "", err
    }

    return tokenString, nil
}




