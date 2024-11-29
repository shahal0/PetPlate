package controllers

import (
	"errors"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)
func AdminLogin(c *gin.Context){
	var loginRequest models.AdminLoginRequest
    if err := c.BindJSON(&loginRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "invalid request data",
        })
        return
    }
	var admin models.Admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		errors.New("password can't be hashed")
	}
	admin.Password = string(hashedPassword)
    tx := database.DB.Where("email = ?", loginRequest.Email).First(&admin)
    if tx.Error != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "invalid email  or password",
        })
        return
    }
	if admin.Password != loginRequest.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "invalid email 11or password",
		})
		return
	}
	tokenString, err := utils.GenerateJWT(admin.Email,"admin")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "failed to generate token",
        })
        return
    }
	c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "login successful",
        "token":   tokenString,
    })
}