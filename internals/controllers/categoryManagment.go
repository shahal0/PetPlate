package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
) 

func CreateCategory(c *gin.Context) {
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
    var req models.CreateCategoryRequest
    // Bind and validate JSON input using the request struct
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
    validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
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
    email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	// Type assertion to string
	_, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve email from token",
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

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Category deleted successfully, products updated to NULL",
    })
}








