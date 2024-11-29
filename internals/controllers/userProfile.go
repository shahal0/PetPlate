package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func GetUserProfile(c *gin.Context) {

	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	emailStr, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	// Check user info and save it into the struct
	var UserProfile models.User
	if err := database.DB.Where("email = ?", emailStr).First(&UserProfile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "User not found",
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
			"Wallet_amount":UserProfile.WalletAmount,
		},
	})
}
func EditProfile(c *gin.Context) {
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	emailStr, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	

	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", emailStr).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "User not found",
		})
		return
	}

	// Perform the update
	if err := database.DB.Model(&user).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Could not edit profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile edited successfully",
		"data":    user,
	})
}
func ChangePassword(c *gin.Context) {
    // Retrieve email from the token
    email, exist := c.Get("email")
    if !exist {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "Unauthorized or invalid token",
        })
        return
    }

    emailStr, ok := email.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to retrieve email from token",
        })
        return
    }

    // Find the user by email
    var user models.User
    if err := database.DB.Where("email = ?", emailStr).First(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "User not found",
        })
        return
    }

    // Bind the request data
    var req models.PasswordChange
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid input",
        })
        return
    }

    // Validate the request data
    validate := validator.New()
    if err := validate.Struct(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": err.Error(),
        })
        return
    }

    // Check if the current password is correct
    err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.OldPassword))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "Current password is incorrect",
        })
        return
    }

    // Hash the new password
    hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to change password, please try again later",
        })
        return
    }

    // Update the user's password in the database
    user.HashedPassword = string(hashedNewPassword)
    if err := database.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to change password, please try again later",
        })
        return
    }

    // Return success response
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Password changed successfully",
    })
}
func WalletHistory(c *gin.Context){
	email, exist := c.Get("email")
    if !exist {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "Unauthorized or invalid token",
        })
        return
    }

    _, ok := email.(string)
	if !ok {
		
	}
	var wh []models.UserWallet
	if err:=database.DB.Model(&models.UserWallet{}).Find(&wh).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":"failed",
			"message":"failed to fetch wallet informations",
		})
		return
	}
	var wr []models.WalletResponse
	for _, wallet := range wh {

		wr=append(wr,models.WalletResponse{
			Amount:wallet.Amount,
			WalletPaymentId: wallet.WalletPaymentId,
			Type: wallet.TypeOfPayment,
			OrderId: wallet.OrderId,
			TransactionTime: wallet.TransactionTime,
			CurrentBalance: wallet.CurrentBalance,
			Reason: wallet.Reason,
		})
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"data":wr,
	})

}