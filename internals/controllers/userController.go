package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"strconv"

	"github.com/gin-gonic/gin"
	//"github.com/go-playground/validator/v10"
	//"golang.org/x/crypto/bcrypt"
)

func UserIDfromEmail(Email string) (ID uint, ok bool) {
	var User models.User
	if err := database.DB.Where("email = ?", Email).First(&User).Error; err != nil {
		return User.ID, false
	}
	return User.ID, true
}

func GetUserList(c *gin.Context) {
	adminid, exist := c.Get("adminID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	_, ok := adminid.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	// Verify the role is "admin"
	var users []models.UserResponse
	tx := database.DB.Model(&models.User{}).Select("id, name, email, phone_number, picture, referral_code, login_method, blocked").Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved user informations",
		"data": gin.H{
			"users": users,
		},
	})
}
func BlockUser(c *gin.Context) {
	adminid, exist := c.Get("adminID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	_, ok := adminid.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	// Get user ID from URL parameter
	userID := c.Query("id")

	// Find the user by ID
	var user models.User

	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid user ID",
		})
		return
	}

	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "user not found",
		})
		return
	}

	// Block the user
	user.Blocked = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to block user",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "user blocked successfully",
	})
}
func UnblockUser(c *gin.Context) {
	adminid, exist := c.Get("adminID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	_, ok := adminid.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	// Get user ID from URL parameter
	userID := c.Query("id")

	var user models.User
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid user ID",
		})
		return
	}

	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "user not found",
		})
		return
	}

	// Unblock the user
	user.Blocked = false
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to unblock user",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "user unblocked successfully",
	})
}




