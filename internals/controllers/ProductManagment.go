package controllers

import (
	"log"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

func AddProducts(c *gin.Context) {
    // Check admin authorization (JWT)
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
    var preq models.AddProductRequest
    if err := c.ShouldBindJSON(&preq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": "failed",
            "message": "invalid input",
        })
        return
    }
    validate := validator.New()
	if err := validate.Struct(&preq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
    if preq.Price<0{
        c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "price should be positive",
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
    validate := validator.New()
	if err := validate.Struct(&preq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
    if preq.Price<0{
        c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "price should be positive",
		})
		return
    }
	if err := database.DB.Model(&product).Updates(preq).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not find product",
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
func ProductRating(c *gin.Context){
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
        var req models.RatingRequest
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
    var ord []models.Order
    if err:=database.DB.Where("user_id=?",userid).Find(&ord).Error;err!=nil{
        c.JSON(http.StatusInternalServerError,gin.H{
            "status":"failed",
            "message":"Error occurred while retrieving orders",
        })
        return
    }
    var oritem  []models.OrderItem
    for _,v:=range ord{
        if err:=database.DB.Where("order_id=?",v.OrderID).Find(&oritem).Error;err!=nil{
            c.JSON(http.StatusInternalServerError,gin.H{
                "status":"failed",
                "message":"Error occurred while retrieving order items",
                })
                return
        }
    }
    flag:=false
    for _,v:=range oritem{
        if v.ProductID==req.ProductID{
            flag=true
        }
    }
    if flag==false{
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "product not found",
        })
        return  
    }
    rate:=models.Rating{
        ProductID: req.ProductID,
        Rating: req.Rating,
        Comment: req.Comment,
    }
    if err:=database.DB.Create(&rate).Error;err!=nil{
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Failed to create rating",
            })
        return    
    }
    c.JSON(http.StatusOK,gin.H{
        "status":"success",
        "message":"Thank you for rating",
    })
}