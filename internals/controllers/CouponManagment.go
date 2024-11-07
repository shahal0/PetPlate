package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
func AdminAddCoupon(c *gin.Context){
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	// Type assertion to string, not uint
	_, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}

    // Bind the incoming JSON to the AddProductRequest struct
    var req models.CouponRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": "failed",
            "message": "invalid input",
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
	if req.DiscountPercentage>100{
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Discount percentage cannot be greater than 100",
			})
			return
	}
	expirationDate := time.Now().AddDate(0, 0, int(req.ExpirationDate))
	coupon:=models.Coupon{
		Code: req.Code,
		DiscountPercentage: req.DiscountPercentage,
		ExpirationDate: expirationDate,
		IsActive:req.IsActive,
		MinimumPurchase: req.MinimumPurchase,
		MaximumDiscountAmount: req.MaximumDiscountAmount,
	}
	if err:=database.DB.Create(&coupon).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create coupon",
			})
			return
	}
	c.JSON(http.StatusOK,gin.H{
		"status": "success",
		"message":"coupon created successfully",
	})
}
func DisableCoupon(c *gin.Context){
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	// Type assertion to string, not uint
	_, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	Couponcode := c.Query("code")
    if Couponcode == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Category ID is required",
        })
        return
    }
	var coupon models.Coupon
	if err:=database.DB.Model(&coupon).Where("code=?",Couponcode).Update("is_active",false).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":"failed",
			"message":"failed to deactivate",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status": "success",
		"message": "coupon deactivated successfully",
	})
}
func ListCoupon(c *gin.Context){
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	// Type assertion to string, not uint
	_, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	var coupons []models.Coupon
	if err:=database.DB.Find(&coupons).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status": "failed",
			"message":"unable to find coupon records",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"data":coupons,
	})
}
