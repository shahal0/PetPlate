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

func AddProducts(c *gin.Context) {
    // Check admin authorization (JWT)
    _, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
        })
        return
    }

    // Bind the incoming JSON to the AddProductRequest struct
    var preq models.AddProductRequest
    if err := c.ShouldBindJSON(&preq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": "failed",
            "message": "invalid input",
        })
        return
    }

    // Create the product object
    product := models.Product{
        Name:        preq.Name,
        Description: preq.Description,
        CategoryID:  preq.CategoryID,
        Price:       preq.Price,
        MaxStock:    preq.MaxStock,
        ImageURL:    preq.ImageURL,
    }

    // Insert the product into the database and check for errors
    if err := database.DB.Create(&product).Error; err != nil {  // <- Fixed error checking
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "could not create product",
        })
        return
    }

    // Success response
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "product added successfully",
        "data":    product,
    })
}
func EditProduct(c *gin.Context){
	_, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
        })
        return
    }
	productID:=c.Query("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Service ID is required",
		})
		return
	}
	var product models.Product
	if err := database.DB.Where("id=?",productID).First(&product).Error;err!=nil{  // <- Fixed error checking
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "could not edit product",
        })
        return
	}
	var preq models.AddProductRequest
	if err := c.ShouldBindJSON(&preq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": "failed",
            "message": "invalid input",
        })
        return
    }
	if err := database.DB.Model(&product).Updates(preq).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not edit category",
        })
        return
    }

	c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "product edited successfully",
        "data":    product,
    })
}
func DeleteProducts(c *gin.Context){
	_, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
        })
        return
    }
	productID:=c.Query("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Service ID is required",
		})
		return
	}
	var product models.Product
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "product not found",
        })
        return
    }
	if err := database.DB.Where("id=?",productID).Delete(&product).Error;err!=nil{  // <- Fixed error checking
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "could not delete product",
        })
        return
	}
	c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "product deleted successfully",
    })

}
func ListProduct(c *gin.Context){
	var product []models.Product
	if err := database.DB.Find(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not retrieve categories",
        })
        return
    }
	c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "product list ",
		"data":product,
    })



}
func CategoryByProduct(c *gin.Context){
    categoryID := c.Query("id")
    if categoryID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Category ID is required",
        })
        return
    }

    // Convert categoryID from string to uint
    parsedID, err := strconv.ParseUint(categoryID, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid Category ID",
        })
        return
    }
    categoryId := uint(parsedID)
    log.Println("Parsed Category ID:", categoryId)
    var product []models.Product
    if err:=database.DB.Where("category_id=?",categoryId).Find(&product).Error;err!=nil{
        log.Println("Error querying products:", err)
        c.JSON(http.StatusNotFound,gin.H{
            "status":"failed",
            "message":"Error occurred while retrieving products",
        })
        return 
    }
    if len(product) == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "No products found for this category",
        })
        return
    }
    c.JSON(http.StatusOK,gin.H{
        "status":"success",
        "message":"product found with category id",
        "data":product,
    })
}