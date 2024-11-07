package controllers

import (
	// "errors"
	//"log"
	"log"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	// "gorm.io/gorm"
)

func PlaceOrder(c *gin.Context) {
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

	// Bind request body to CartRequest struct
	couponcode := c.Query("couponCode")
	var req models.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}
	var paymentMethod string
	switch req.PaymentMethod {
	case 1:
		paymentMethod = "COD"
	case 2:
		paymentMethod = "UPI"
	case 3:
		paymentMethod = "Wallet"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid payment method",
		})
		return
	}

	// Validate the struct
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	var ucart []models.Cart
	if err := database.DB.Where("user_id=?", userid).Find(&ucart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	var address models.Address
	if err := database.DB.Where("address_id=?", req.AddressID).First(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Invalid address id",
		})
		return
	}
	var message  string
	TotalAmount,Rawamount:= utils.Itercart(ucart)
	var coup models.Coupon
	if couponcode!=""{
	 database.DB.Where("code=?", couponcode).First(&coup)
	 message="coupon is applied"
		
	var coupus  models.CouponUsage
	 database.DB.Where("coupon=?",couponcode).First(&coupus)
	 if coup.IsActive==false{
		message="coupon is not active"
	}else if coupus.UsageCOunt>=3{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Coupon usage limit is over",
			})
			return
	}else if coup.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Coupon is expired",
			"coupon Status":message,
		})
		return
	}
	log.Println("total amount",TotalAmount)
	if TotalAmount<coup.MinimumPurchase{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":  "failed",
			"message": "Minimum purchase amount is not met",
		})
		return
	}else if coup.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
		message="Coupon is expired"
	}
	}else{
		message="no coupon applied"
	}
	
	discountAmount := TotalAmount * (coup.DiscountPercentage / 100)
	if discountAmount>coup.MaximumDiscountAmount{
		discountAmount=coup.MaximumDiscountAmount
	}
	order := models.Order{
		UserID:        userid,
		OrderDate:     time.Now(),
		RawAmount: Rawamount,
		OfferTotal:         TotalAmount,
		CouponCode: couponcode,
		DiscountAmount: discountAmount,
		FinalAmount: TotalAmount-discountAmount,
		PaymentMethod: paymentMethod,
		PaymentStatus: models.Pending,
		OrderStatus:   models.Pending,
		ShippingAddress: models.ShippingAddress{
			PhoneNumber:  address.PhoneNumber,
			StreetName:   address.StreetName,
			StreetNumber: address.StreetNumber,
			City:         address.City,
			State:        address.State,
			PostalCode:   address.PostalCode,
		},
	}
	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	utils.CartToOrderItem(*c, ucart, order.OrderID)
	var cu models. CouponUsage
	if err := database.DB.Model(&cu).Where("coupon = ? AND user_id = ?", couponcode, userid).Update("usage_count",cu.UsageCOunt+1).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}else{
		couponusage:=models.CouponUsage{
			Coupon: couponcode,
			UserID: userid,
			UsageCOunt: 1,
		}
		if err:=database.DB.Create(&couponusage).Error;err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order created successfully",
		"data":    order,
		"coupon Status":message,
	})
}
func UserSeeOrders(c *gin.Context) {
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
	var ord []models.Order
	if err := database.DB.Where("user_id", userid).Find(&ord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	var ordItems []models.OrderItem
	var ordresponse []models.OrderResponse

	for _, ords := range ord {
		if err := database.DB.Where("order_id", ords.OrderID).Find(&ordItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
		var orderitems []models.OrderItemResponse
		for _, ordItem := range ordItems {
			var prodct models.Product
			if err := database.DB.Where("product_id=?", ordItem.ProductID).First(&prodct).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": err.Error(),
				})
				return
			}
			orderitems = append(orderitems, models.OrderItemResponse{
				ProductId:   ordItem.ProductID,
				OrderID:     ordItem.OrderID,
				ProductName: prodct.Name,
				ImageURL:    prodct.ImageURL,
				CategoryId:  prodct.CategoryID,
				Description: prodct.Description,
				Price:       prodct.OfferPrice,
				Quantity:    ordItem.Quantity,
				TotalPrice:  float64(ordItem.Quantity) * prodct.OfferPrice,
				OrderStatus: ordItem.OrderStatus,
			})
		}
		ordresponse = append(ordresponse, models.OrderResponse{
			OrderID:         ords.OrderID,
			OrderDate:       ords.OrderDate,
			RawAmount: ords.RawAmount,
			OfferTotal: ords.OfferTotal,
			DiscountPrice: ords.DiscountAmount,
			FinalAmount: ords.FinalAmount,
			ShippingAddress: ords.ShippingAddress,
			OrderStatus:     ords.OrderStatus,
			PaymentStatus:   ords.PaymentStatus,
			PaymentMethod:   ords.PaymentMethod,
			Items:           orderitems,
		})

	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   ordresponse,
	})

}
func UserCancelOrder(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Unauthorized or invalid token"})
		return
	}
	ordid := c.Query("id")
	if ordid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID is required"})
		return
	}

	parsedID, err := strconv.ParseUint(ordid, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid Order ID"})
		return
	}
	orderID := uint(parsedID)

	if err := utils.CancelOrder(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order cancelled successfully"})
}
func CancelItemFromUserOrders(c *gin.Context) {
	if _, exist := c.Get("email"); !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Unauthorized or invalid token"})
		return
	}

	ordid := c.Query("id")
	prodid := c.Query("productid")
	if ordid == "" || prodid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID and Product ID are required"})
		return
	}

	// Convert IDs to uint
	parsedOrderID, err := strconv.ParseUint(ordid, 10, 32)
	parsedProductID, errProd := strconv.ParseUint(prodid, 10, 32)
	if err != nil || errProd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid Order ID or Product ID"})
		return
	}

	// Cancel the order item
	if err := utils.CancelOrderItem(uint(parsedOrderID), uint(parsedProductID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order item cancelled and order total updated successfully"})
}
func AdminOrderList(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	var orders []models.Order
	if err := database.DB.Model(&models.Order{}).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve orders",
		})
		return
	}
	var ordresponses []models.AdminOrderResponse
	for _, order := range orders {
		var orditemresp []models.OrderItemResponse
		var orderitem []models.OrderItem
		if err := database.DB.Model(&orderitem).Where("order_id=?", order.OrderID).Find(&orderitem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to retrieve order items",
			})
			return
		}
		for _, orditem := range orderitem {
			var prodct models.Product
			if err := database.DB.Model(&prodct).Where("id=?", orditem.ProductID).First(&prodct).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "Failed to retrieve product",
				})
				return
			}
			orditemresp = append(orditemresp, models.OrderItemResponse{
				OrderID:     orditem.OrderID,
				ProductId:   orditem.ProductID,
				ProductName: prodct.Name,
				ImageURL:    prodct.ImageURL,
				CategoryId:  prodct.CategoryID,
				Description: prodct.Description,
				Price:       prodct.OfferPrice,
				Quantity:    orditem.Quantity,
				TotalPrice:  prodct.Price * float64(orditem.Quantity),
				OrderStatus: orditem.OrderStatus,
			})
		}
		var ussr models.User
		if err := database.DB.Model(&ussr).Where("id=?", order.UserID).First(&ussr).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to retrieve user",
			})
			return
		}
		ordresponses = append(ordresponses, models.AdminOrderResponse{
			UserId:          order.UserID,
			UserName:        ussr.Name,
			OrderID:         order.OrderID,
			OrderDate:       order.OrderDate,
			OfferTotal: order.OfferTotal,
			DiscountPrice: order.DiscountAmount,
			FinalAmount: order.FinalAmount,
			ShippingAddress: order.ShippingAddress,
			OrderStatus:     order.OrderStatus,
			PaymentStatus:   order.PaymentStatus,
			PaymentMethod:   order.PaymentMethod,
			Items:           orditemresp,
		})

	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   ordresponses,
	})

}
func AdminCancelOrder(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	ordid := c.Query("id")
	if ordid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID is required"})
		return
	}

	parsedID, err := strconv.ParseUint(ordid, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid Order ID"})
		return
	}
	orderID := uint(parsedID)

	if err := utils.CancelOrder(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order cancelled successfully"})
}

func CancelItemFromAdminOrders(c *gin.Context) {
	ordid := c.Query("id")
	prodid := c.Query("productid")
	if ordid == "" || prodid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID and Product ID are required"})
		return
	}

	// Convert IDs to uint
	parsedOrderID, err := strconv.ParseUint(ordid, 10, 32)
	parsedProductID, errProd := strconv.ParseUint(prodid, 10, 32)
	if err != nil || errProd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid Order ID or Product ID"})
		return
	}

	// Cancel the order item
	if err := utils.CancelOrderItem(uint(parsedOrderID), uint(parsedProductID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order item cancelled and order total updated successfully"})
}
func UpdateOrderstatus(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}
	ordid := c.Query("id")
	if ordid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID is required"})
		return
	}

	parsedID, err := strconv.ParseUint(ordid, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid Order ID"})
		return
	}
	orderID := uint(parsedID)

	var ord models.Order
	if err := database.DB.Where("order_id=?", orderID).First(&ord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order",
		})
		return
	}
	var orditems []models.OrderItem
	if err := database.DB.Where("order_id=?", orderID).Find(&orditems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrive order items",
		})
		return
	}
	var updatedSatus string
	if ord.OrderStatus == models.Pending {
		updatedSatus = models.Confirm
		for _, item := range orditems {
			if item.OrderStatus != models.Cancelled {
				if err := database.DB.Model(&models.OrderItem{}).Where("order_id = ?", item.OrderID).Update("order_status", updatedSatus).Error; err != nil {
					c.JSON(http.StatusNotModified, gin.H{
						"status":  "failed",
						"message": "unable to update orderitem status",
					})
					return
				}
			}
		}
	} else if ord.OrderStatus == models.Confirm {
		updatedSatus = models.Shipped
		for _, item := range orditems {
			if item.OrderStatus != models.Cancelled {
				if err := database.DB.Model(&models.OrderItem{}).Where("order_id = ?", item.OrderID).Update("order_status", updatedSatus).Error; err != nil {
					c.JSON(http.StatusNotModified, gin.H{
						"status":  "failed",
						"message": "unable to update orderitem status",
					})
					return
				}
			}
		}
	} else {
		updatedSatus = models.Delivered
		for _, item := range orditems {
			if item.OrderStatus != models.Cancelled {
				if err := database.DB.Model(&models.OrderItem{}).Where("order_id = ?", item.OrderID).Update("order_status", updatedSatus).Error; err != nil {
					c.JSON(http.StatusNotModified, gin.H{
						"status":  "failed",
						"message": "unable to update orderitem status",
					})
					return
				}
			}
		}
	}
	if err := database.DB.Model(&models.Order{}).Where("order_id = ?", ord.OrderID).Update("order_status", updatedSatus).Error; err != nil {
		c.JSON(http.StatusNotModified, gin.H{
			"status":  "failed",
			"message": "unable to update orderitem status",
		})
		return
	}
	if ord.OrderStatus == models.Delivered {
		ord.PaymentStatus = models.Success
	}
	database.DB.Save(&ord)
	c.JSON(http.StatusOK, gin.H{
		"status":                   "success",
		"message":                  "order status updated successfully",
		"your order status is now": updatedSatus,
	})

}
