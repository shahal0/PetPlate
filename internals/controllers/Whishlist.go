package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
func AddToWhishlist(c *gin.Context) {
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
	var req models.WhislistRequest
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
	var product models.Product
		if err := database.DB.Model(&models.Product{}).Where("product_id = ?", req.ProductID).First(&product).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "unable find the product",
			})
			return
		}
	var whislist models.Whishlist
	err:=database.DB.Model(&whislist).Where("user_id=? AND product_id=?",userid,req.ProductID).First(&whislist).Error
	if err!=nil{
	whishlist:=models.Whishlist{
		UserID: userid,
		ProductID: req.ProductID,
	}
	if err := database.DB.Create(&whishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Could not add to whishlist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "product added to whishlist successfully",
	})
		return
	}

		c.JSON(http.StatusBadRequest,gin.H{
			"status":"failed",
			"message":"product already in whishlist",
		})
}
func SowWhislist(c *gin.Context){
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
	userid, ok := utils.GetUserIDByEmail(emailval)
	var whishlists []models.Whishlist
	if err:=database.DB.Model(&whishlists).Where("user_id=?",userid).Find(&whishlists).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve whishlist",
		})
		return
	}
	var response []models.ResponseWhishlist
	for _,wl:=range whishlists{
		var prod models.Product
		if err:=database.DB.Model(&prod).Where("product_id=?",wl.ProductID).First(&prod).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"status":  "failed",
				"message": "Failed to retrieve product",
			})
			return
		}
		response=append(response,models.ResponseWhishlist{
			ProductID:wl.ProductID,
			ProductName:prod.Name,
			ProductImage: prod.ImageURL,
			ProductDescription: prod.Description,
			ProductPrice: prod.Price,
			OfferPrice: prod.OfferPrice,

		})

	}
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"data": response,
	})

}
func AddToCart(c *gin.Context){
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
	userid, ok := utils.GetUserIDByEmail(emailval)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
			})
			return
	}
	productId:=c.Query("productid")
	num, errr := strconv.ParseUint(productId, 10, 32) // base 10, 32-bit size
    if errr != nil {
        c.JSON(http.StatusBadRequest,gin.H{
			"status":  "failed",
			"message": "Invalid product id",
		})
		return
    }

    // Type cast to uint
    uintproductid:= uint(num)
	var product models.Product
	if err:=database.DB.Model(&product).Where("product_id=?",productId).First(&product).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve product",
		})
		return
	}
	var cart models.Cart
	err:=database.DB.Model(&models.Cart{}).Where("user_id=? AND product_id=?",userid,productId).First(&cart).Error

	 if err!=nil{
		cart:=models.Cart{
			UserID:userid,
			ProductID:uintproductid,
			Quantity: 1,
		}
		if err:=database.DB.Create(&cart).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"status":  "failed",
				"message": "Failed to add product to cart",
				})
				return
		}
		c.JSON(http.StatusOK,gin.H{
			"status":  "success",
			"message": "Product added to cart",
		})
		return
	 }
	 var cartt models.Cart
	 if err:=database.DB.Model(&cartt).Where("user_id=? AND product_id=?",userid,productId).Update("quantity",cartt.Quantity+1).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to update cart",
			})
			return
	 }
	 var wishlist models.Whishlist
	if err := database.DB.Where("product_id = ?", productId).Delete(&wishlist).Error; err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  "failed",
		"message": "Failed to delete wishlist item",
	})
	return
	}
	 c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"message": "Product quantity updated",
	 })
	 
	 
}