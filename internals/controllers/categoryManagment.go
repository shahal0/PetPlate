package controllers

import (
    
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/gin-gonic/gin"

) 

func CreateCategory(c *gin.Context) {
	_, err := utils.GetJWTClaim(c) // Assume this is the admin email
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized or invalid token",
		})
		return
	}
    var req models.CreateCategoryRequest
    // Bind and validate JSON input using the request struct
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid input",
        })
        return
    }

    // Map the request to the Category model
    category := models.Category{
        Name:        req.Name,
        Description: req.Description,
    }

    // Insert the category into the database
    if err := database.DB.Create(&category).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not create category",
        })
        return
    }

    // Respond with success and return the created category
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Category created successfully",
        "data":    category,
    })
}
func GetCategories(c *gin.Context){
	var categories []models.Category
	if err := database.DB.Find(&categories).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not retrieve categories",
        })
        return
    }
	// Send the categories in the response
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   categories,
    })
}
func CategoryEdit(c *gin.Context) {
    // Check authorization (admin email assumed from JWT claim)
    _, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
        })
        return
    }

    // Get category ID from the query parameter
    categoryID := c.Query("id")
    if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Service ID is required",
		})
		return
	}
    var category models.Category
    // Find the category by ID
    if err := database.DB.Where("id = ?", categoryID).First(&category).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "Category not found",
        })
        return
    }
    var req models.CreateCategoryRequest
    

    // Bind JSON input to the category object (updating fields like name, description)
    if err := c.ShouldBindJSON(&category); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid input",
        })
        return
    }

    // Save the updated category
    if err := database.DB.Model(&category).Updates(req).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not edit category",
        })
        return
    }

    // Success response
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Category updated successfully",
        "data":    category,
    })
}
func CategoryDelete(c *gin.Context) {
    // Extract and verify JWT claim
    _, err := utils.GetJWTClaim(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "Unauthorized or invalid token",
        })
        return
    }

    // Get category ID from query params
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

    // Find the category by ID
    var category models.Category
    if err := database.DB.Where("id = ?", categoryId).First(&category).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "Category not found",
        })
        return
    }

    // Delete the category from the database
    if err := database.DB.Delete(&category).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to delete category",
        })
        return
    }

    // Update all products associated with this category's ID to set CategoryID to NULL
    if err := database.DB.Model(&models.Product{}).
        Where("category_id = ?", categoryId).
        Update("category_id", nil).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to update products in the category",
        })
        return
    }

    // Success response after category deletion and product update
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Category deleted successfully, products updated to NULL",
    })
}








