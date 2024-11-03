package controllers

import (
	"log"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
func AddAddress(c *gin.Context){
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

	var count []models.Address
	database.DB.Model(&models.Address{}).Where("user_id=?",userid).Find(&count)
	if len(count)>=3{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":"failed",
			"message":"maximum 3 address allowed",
		})
		return 
	}
	var req models.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": "failed",
            "message": "invalid input",
        })
        return
    }
    validate:= validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	address:=models.Address{
		UserID: userid,
		AddressID: req.AddressID,
		PhoneNumber: req.PhoneNumber,
		AddressType: req.AddressType,
		StreetName: req.StreetName,
		StreetNumber: req.StreetNumber,
		City: req.City,
		State: req.State,
		PostalCode: req.PostalCode,
	}
	if err:=database.DB.Create(&address).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "could not add address",
        })
        return
	}
	c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Address added successfully",
        "data":    address,
    })
}
func EditAddress(c *gin.Context) {
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
	AddressID := c.Query("id")
    if AddressID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Address ID is required",
        })
        return
    }

    // Convert categoryID from string to uint
    parsedID, err := strconv.ParseUint(AddressID, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid Address ID",
        })
        return
    }
    addressId := uint(parsedID)
	var address models.Address
	if err := database.DB.Where("address_id = ?", addressId).First(&address).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "Category not found",
        })
        return
    }
	var req models.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		log.Println("error",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid input",
		})
		return
	}
	if err := database.DB.Model(&address).Updates(req).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "Could not edit Address",
        })
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Address updated successfully",
		"data":address,
	})
}
func DeleteAddress(c *gin.Context){
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
	AddressID := c.Query("id")
    if AddressID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Address ID is required",
        })
        return
    }

    // Convert categoryID from string to uint
    parsedID, err := strconv.ParseUint(AddressID, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "Invalid address ID",
        })
        return
    }
    addressId := uint(parsedID)
	var address models.Address
	if err := database.DB.Where("address_id = ?", addressId).First(&address).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "failed",
            "message": "Address not found",
        })
        return
    }
	if err := database.DB.Where("address_id=?",addressId).Delete(&address).Error;err!=nil{ 
		log.Println("error",err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "could not delete address",
        })
        return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"message":"address deleted succesfully",
	})
}