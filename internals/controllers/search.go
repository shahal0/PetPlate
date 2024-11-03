package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
)
func SearchProduct(c *gin.Context){
	if _, exist := c.Get("email"); !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Unauthorized or invalid token"})
		return
	}
	category_id:=c.Query("category_id")
	sortBy:=c.Query("sort_by")

	var products []models.Product
	var rating []models.Rating
	qry:=database.DB.Model(&rating)
	Query:=database.DB.Model(&products)
	Query=Query.Where("category_id=? AND max_stock>0",category_id,).Find(&products)
	switch(sortBy){
	case "price_H-L":
		Query.Order("price desc")
	case  "price_L-H":
		Query.Order("price asc")
	case  "newest":
		Query.Order("created_at desc")
	case "alphabetic":
	 Query.Order("Lower(name) ASC")
	case "popularity":
		qry.Order("rating desc")	
	}
	if err:=Query.Find(&products).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
	}
	c.JSON(http.StatusOK,gin.H{
		"status": "success",
		"data": products,
	})

}