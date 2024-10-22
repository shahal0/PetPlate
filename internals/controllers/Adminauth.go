package controllers

import (
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
            "status":  false,
            "message": "invalid request data",
        })
        return
    }
	var admin models.Admin
    tx := database.DB.Where("email = ?", loginRequest.Email).First(&admin)
    if tx.Error != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  false,
            "message": "invalid email or password",
        })
        return
    }
	err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(loginRequest.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  false,
            "message": "invalid email or password",
        })
        return
    }
	tokenString, err := utils.GenerateJWT(admin.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  false,
            "message": "failed to generate token",
        })
        return
    }
	c.JSON(http.StatusOK, gin.H{
        "status": true,
        "message": "login successful",
        "token":   tokenString,
    })
}