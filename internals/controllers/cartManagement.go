package controllers

import (
	"errors"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)
func AddCart(c *gin.Context) {
	// Retrieve email from the context
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	emailval, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}

	// Retrieve user ID based on email
	userid, ok := utils.GetUserIDByEmail(emailval)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}
	var req models.CartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}

	// Validate the struct
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	// Check if product is already in cart
	var existingCart models.Cart
	database.DB.Where("user_id = ? AND product_id = ?", userid, req.ProductID).First(&existingCart)
	if existingCart.ProductID==req.ProductID{
		newQuantity:=existingCart.Quantity+req.Quantity
		var product models.Product
		if err:=database.DB.Where("product_id=?",req.ProductID).First(&product).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to retrieve product from database",
				})
				return
		}
		if newQuantity>product.MaxStock{
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Quantity exceeds product stock",
				})
				return
		}
		if newQuantity>10{
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Quantity exceeds maximum allowed",
				})
				return
		}
		if err := database.DB.Model(&existingCart).Where("user_id=? and product_id=?", userid, req.ProductID).Update("quantity", newQuantity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK,gin.H{
			"status": "success",
			"message": "Product quantity updated successfully",
		})
		return

	}

	// Product not in cart, create new cart entry
	if req.Quantity > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Maximum quantity allowed is 10",
		})
		return
	}

	cart := models.Cart{
		UserID:    userid,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	if err := database.DB.Create(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Could not add to cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cart added successfully",
	})
}

func ListCart(c *gin.Context){
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	emailval, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	couponcode:=c.Query("couponcode")
	userid,ok:=utils.GetUserIDByEmail(emailval)
	if !ok{
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"message": "invalid input",
		})
		return
	}
	var cart []models.Cart
	if err:=database.DB.Model(models.Cart{}).Where("user_id=?",userid).Find(&cart).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message":"cart not found",
		})
		return
	}
	var message string
	var TotalAmount float64
	flag:=0
	var coupon  models.Coupon
	var responseCarts[]models.CartProduct
	for _,cartitem:=range cart{
		var product models.Product
		result := database.DB.Model(models.Product{}).Where("product_id=?", cartitem.ProductID).First(&product)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  "failed",
					"message": "product not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "internal server error",
			})
			return
		}
		database.DB.Model(&models.Coupon{}).Where("code=?",couponcode).First(&coupon)
		if couponcode==""{
			message="no coupon applied"
		}else if !coupon.IsActive{
			message="coupon is not active"
		}else{
			message="coupon applied"
			flag=1
		}
		responseCart:=models.CartProduct{
			Product_id:cartitem.ProductID,
			Name:product.Name,
			Description:product.Description,
			CategoryID:product.CategoryID,
			Price  :product.Price,
			FinalPrice: product.OfferPrice,
			Quantity: cartitem.Quantity,
			ImageURL :product.ImageURL,
		}
		var category models.Category
		database.DB.Model(&category).Where("id=?",product.CategoryID).First(&category)
		responseCarts=append(responseCarts,responseCart)
		amount:=product.OfferPrice*float64(cartitem.Quantity)
		amount-=product.Price*(category.CategoryOffer/100)	
		TotalAmount+=amount
	}
	if flag==1{
		if TotalAmount<coupon.MinimumPurchase{
			c.JSON(http.StatusBadRequest,gin.H{
				"status":  "failed",
				"message": "minimum purchase amount not met",
			})
			return
			}
		TotalAmount-=TotalAmount*(coupon.DiscountPercentage/100)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    responseCarts,
		"CartAmount":RoundDecimalValue(TotalAmount),
		"message":message,
	})
	}
func DeleteFromCart(c *gin.Context){
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
		})
		return
	}
	product := c.Query("id")
    if product == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Category ID is required",
        })
        return
    }

    // Convert categoryID from string to uint
    parsedID, err := strconv.ParseUint(product, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid Category ID",
        })
        return
    }
    productId := uint(parsedID)
	var cartt models.Cart
	if err:=database.DB.Where("product_id=?",productId).Delete(&cartt).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to delete product from cart",
			})
			return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"message":"product deleted from the cart",
	})
}

