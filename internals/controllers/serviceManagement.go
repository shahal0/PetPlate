package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"

	"github.com/gin-gonic/gin"
)

func AddService(c *gin.Context){
	_, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
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
	_,err:=utils.GetJWTClaim(c)
	if err!=nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"status":"failed",
			"message":"unauthorized or invalid Token",
		})
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
	_, err := utils.GetJWTClaim(c) 
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "failed",
            "message": "unauthorized or invalid token",
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