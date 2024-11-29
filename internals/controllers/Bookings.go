package controllers

import (
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"

	// "strconv"
	"time"

	"github.com/gin-gonic/gin"
)
func Booking(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}
	payment_method:=""
	switch req.PaymentMethod {
		case 1:
			payment_method = "RayzorPay" 
		case 2:
			payment_method = "Wallet"
			default:
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "failed",
					"message": "Invalid payment method",
					})
					return
		}

	if req.TimeSlot=="" || req.Date==""||req.ServiceId<1{
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
			})
			return
	}
	parsedDate, err := time.Parse(models.Layout, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid date format",
			})
		return
	}
	var service models.Service
	database.DB.Where("id=?",req.ServiceId).First(&service)
	var existingBooking models.Booking
	errr := database.DB.Where("service_id = ? AND booking_date = ? AND time_slot = ?", req.ServiceId, parsedDate, req.TimeSlot).First(&existingBooking).Error
	if errr == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Time slot already booked",
		})
		return
	} 
	booking:=models.Booking{
		UserId: userid,
		ServiceID: req.ServiceId,
		TimeSlot: req.TimeSlot,
		BookingDate: parsedDate,
		BookingStatus: models.Pending,
		Amount: service.Price,
		PaymentMethod: payment_method,
		PaymentStatus: models.Pending,
	}
	if err:=database.DB.Create(&booking).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to create booking",
		})
		return
	}
	if payment_method==models.Wallet{
		if err:=utils.WalletPayment(userid,booking.BookingID);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"message": "Booking created successfully",
		})	
}
func UserGetBooking(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}
	var bookings []models.Booking
	if err:=database.DB.Model(&models.Booking{}).Where("user_id = ?", userid).Find(&bookings).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve bookings",
		})
	}
	var resbookings []models.BookingResponse
	for _, booking := range bookings{
		var service models.Service
		if err:=database.DB.Where("id=?",booking.ServiceID).First(&service).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"status":  "failed",
				"message": "Failed to retrieve service",
			})
			return
		}
		resbookings=append(resbookings, models.BookingResponse{
			UserID: userid,
			BookingId: booking.BookingID,
			Service: service.Name,
			TimeSlot: booking.TimeSlot,
			BookingDate: booking.BookingDate,
			BookingStatus: booking.BookingStatus,
			Amount: booking.Amount,
			PaymentMethod: booking.PaymentMethod,
			PaymentStatus: booking.BookingStatus,
		})
	}
	 c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data": resbookings,
		})
}
func GetBookingAdmin(c *gin.Context) {
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
	var bookings []models.Booking
	if err:=database.DB.Model(&models.Booking{}).Find(&bookings).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve bookings",
		})
		return
	}
	var resbookings []models.BookingResponse
	for _, booking := range bookings{
		var service models.Service
		if err:=database.DB.Where("service_id=?",booking.ServiceID).First(&service).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"status":  "failed",
				"message": "Failed to retrieve service",
			})
			return
		}
		resbookings=append(resbookings, models.BookingResponse{
			UserID: booking.UserId,
			BookingId: booking.BookingID,
			Service: service.Name,
			TimeSlot: booking.TimeSlot,
			BookingDate: booking.BookingDate,
			BookingStatus: booking.BookingStatus,
			Amount: booking.Amount,
			PaymentMethod: booking.PaymentMethod,
			PaymentStatus: booking.BookingStatus,
		})
	}
	 c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data": resbookings,
		})
}