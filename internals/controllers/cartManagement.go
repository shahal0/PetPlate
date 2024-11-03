package controllers

import (
	"errors"
	"log"
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
	err := database.DB.Where("user_id = ? AND product_id = ?", userid, req.ProductID).First(&existingCart).Error

	if err == nil {
		
		newQuantity := existingCart.Quantity + req.Quantity
		if newQuantity > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Maximum quantity allowed is 10",
			})
			return
		}
		var product models.Product
		if err := database.DB.Model(&models.Product{}).Where("id = ?", req.ProductID).First(&product).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "unable find the product",
			})
			return
		}
		if newQuantity>product.MaxStock{
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Not enough stock",
				})
				return 
		}

		existingCart.Quantity = newQuantity
	
	
		// Use Update instead of Save to ensure the record with the correct ID is updated
		if err := database.DB.Model(&existingCart).Where("user_id = ? AND product_id = ?", userid, req.ProductID).Updates(map[string]interface{}{
			"quantity": existingCart.Quantity,
		}).Error; err != nil {
			log.Println("error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to update cart",
			})
			return
		}
	
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Cart updated successfully",
		})
		return
	}else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Handle errors other than "not found"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to check existing cart",
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
	var TotalAmount float64
	var responseCarts[]models.CartProduct
	for _,cartitem:=range cart{
		var product models.Product
		result := database.DB.Model(models.Product{}).Where("id=?", cartitem.ProductID).First(&product)
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
			log.Printf("error %v", result.Error)
			return
		}
		responseCart:=models.CartProduct{
			Product_id:cartitem.ProductID,
			Name:product.Name,
			Description:product.Description,
			CategoryID:product.CategoryID,
			Price  :product.Price,
			Quantity: cartitem.Quantity,
			ImageURL :product.ImageURL,
		}
		responseCarts=append(responseCarts,responseCart)
		amount:=product.Price*float64(cartitem.Quantity)	
		TotalAmount+=amount
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    responseCarts,
		"CartAmount":TotalAmount,
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
	if err:=database.DB.Delete(&cartt,productId).Error;err!=nil{
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
