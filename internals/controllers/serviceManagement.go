package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddService(c *gin.Context){
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
	var req models.ServiceRequest
	if err:=c.ShouldBind(&req);err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
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
	if req.Price<0{
        c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "price should be positive",
		})
		return
    }
	service:=models.Service{
		ID: req.ID,
		Name: req.Name,
		Description: req.Description,
		Price: req.Price,
		ImageURL: req.ImageURL,
	}
	if err:=database.DB.Create(&service).Error;err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"status": "failed",
			"message": "unable to add services",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "service added successfully",
        "data":    service,
    })
}
func EditService(c *gin.Context){
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
	ServiceID:=c.Query("id")
	if ServiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Service ID is required",
		})
		return
	}
	
	var req models.ServiceRequest
	if err:=c.ShouldBind(&req);err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":"failed",
			"message":"invalid input",
		})
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	if req.Price<0{
        c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "price should be positive",
		})
		return
    }
	var service models.Service
	if err:=database.DB.Where("id = ?",ServiceID).First(&service).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":"failure",
			"message":"could not find record",
		})
	}
	if err := database.DB.Model(&service).Updates(req).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not edit services",
        })
        return
    }

	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"message":"service edited successfully",
		"data":service,
	})
	

}
func DeleteService(c *gin.Context){
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
	ServiceID:=c.Query("id")
	if ServiceID==""{
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Service ID is required",
		})
		return
	}
	
	var service models.Service
	if err:=database.DB.Where("id=?",ServiceID).First(&service).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":"failed",
			"message":"could not find recoed",
		})
		return 
	}
	if err:=database.DB.Where("id=?",ServiceID).Delete(&service).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":"failed",
			"message":"could not delete the service",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"message":"service deleted succesfully",
	})
}
func GetServices(c *gin.Context){
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
	var service []models.Service
	if err:=database.DB.Find(&service).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":"failed",
			"message":"couldnt found any records",
		})
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"message":"service list",
		"data":service,
	})
}