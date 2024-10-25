package controllers

import (
	"log"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)



func UserIDfromEmail(Email string) (ID uint, ok bool) {
	var User models.User
	if err := database.DB.Where("email = ?", Email).First(&User).Error; err != nil {
		return User.ID, false
	}
	return User.ID, true
}
func GetUserProfile(c *gin.Context) {
	// Check user API authentication by extracting the JWT claim (email)
	claims, err := utils.GetJWTClaim(c)  // This returns *jwt.MapClaims
	if err != nil {
		// Handle the error properly
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized or invalid token",
		})
		return
	}

	// Extract the email from the claims
	email, ok := (*claims)["email"].(string) // Make sure to cast it to a string
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "invalid token structure, email claim missing",
		})
		return
	}

	// Assuming you have a function to get the user ID from email
	UserID, ok := UserIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no such user record",
		})
		return
	}

	// Check user info and save it into the struct
	var UserProfile models.User
	if err := database.DB.Where("id = ?", UserID).First(&UserProfile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	// Return the user profile
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully fetched user profile",
		"data": gin.H{
			"id":            UserProfile.ID,
			"name":          UserProfile.Name,
			"email":         UserProfile.Email,
			"phone_number":  UserProfile.PhoneNumber,
			"picture":       UserProfile.Picture,
			"login_method":  UserProfile.LoginMethod,
			"blocked":       UserProfile.Blocked,
			//"wallet_amount": UserProfile.WalletAmount,
		},
	})
}

func GetUserList(c *gin.Context) {
	// 1. Verify the admin's authentication
	log.Println("10")
	_, err := utils.GetJWTClaim(c) // Assume this is the admin email
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized or invalid token",
		})
		return
	}
	var users []models.UserResponse
	tx := database.DB.Model(&models.User{}).Select("id, name, email, phone_number, picture, referral_code, login_method, blocked").Find(&users)
	if tx.Error != nil {
		log.Println("Database query error:", tx.Error)
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
	_, err := utils.GetJWTClaim(c) // Assume this is the admin email
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized or invalid token",
		})
		return
	}
    // Get user ID from URL parameter
    userID := c.Query("id")
	log.Println("user id",userID)

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
	_, err := utils.GetJWTClaim(c) // Assume this is the admin email
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized or invalid token",
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


