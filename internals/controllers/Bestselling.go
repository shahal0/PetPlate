package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
	// "gorm.io/driver/postgres"
	// "gorm.io/gorm"
)
func BestSellingProduct(c *gin.Context) {
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
	var bestProduct []models.BestSellingProduct
	database.DB.Raw(`
	SELECT 
	p.product_id AS product_id,
	p.name AS product_name,
	SUM(o.quantity) AS total_sold
	FROM 
	order_items o
	JOIN 
	products p ON o.product_id = p.product_id
	GROUP BY 
	p.product_id, p.name
	ORDER BY 
	total_sold DESC
	LIMIT 10;
	`).Scan(&bestProduct)
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"BestSelling Product":bestProduct,
	})
}
func BestsellingCategory(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
			})
			return
		}
	var bestCategory []models.BestSellingCategory
	database.DB.Raw(`
	SELECT 
	c.id AS category_id,
	c.name AS category_name,
	SUM(o.quantity) AS total_sold
	FROM 
	order_items o
	JOIN 
	products p ON o.product_id = p.product_id
	JOIN 
	categories c ON p.category_id = c.id
	GROUP BY 
	c.id, c.name
	ORDER BY 
	total_sold DESC
	LIMIT 10;
	`).Scan(&bestCategory)
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"BestSelling category":bestCategory,
	})
}
	
