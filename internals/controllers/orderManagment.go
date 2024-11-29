package controllers

import (
	"fmt"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"petplate/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	//"google.golang.org/protobuf/internal/order"
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
		paymentMethod =models.Wallet
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
	var message string
	TotalAmount, Rawamount := utils.Itercart(ucart)
	
	var coup models.Coupon
	if couponcode != "" {
		database.DB.Where("code=?", couponcode).First(&coup)
		message = "coupon is applied"

		var coupus models.CouponUsage
		database.DB.Where("coupon=?", couponcode).First(&coupus)
		if coup.IsActive == false {
			message = "coupon is not active"
		} else if coupus.UsageCOunt >= 3 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Coupon usage limit is over",
			})
			return
		} else if coup.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":        "failed",
				"message":       "Coupon is expired",
				"coupon Status": message,
			})
			return
		}
		if TotalAmount < coup.MinimumPurchase {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Minimum purchase amount is not met",
			})
			return
		} else if coup.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
			message = "Coupon is expired"
		}
	} else {
		message = "no coupon applied"
	}
	if TotalAmount<coup.MinimumPurchase{
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Minimum purchase amount is not met coupon not applied",
			})
	}
	final,discount:=utils.Distercart(ucart,coup)
	if discount>coup.MaximumDiscountAmount{
		discount=coup.MaximumDiscountAmount
	}
	delivery:=0
	if final<=1000{
		delivery=50
	}
	order := models.Order{
		UserID:         userid,
		OrderDate:      time.Now(),
		RawAmount:      Rawamount,
		OfferTotal:     TotalAmount,
		CouponCode:     couponcode,
		DiscountAmount: discount,
		DeliveryCharge: float64(delivery),
		FinalAmount:    final+float64(delivery),
		PaymentMethod:  paymentMethod,
		PaymentStatus:  models.Pending,
		OrderStatus:    models.Pending,
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
	if req.PaymentMethod==1&&order.FinalAmount>1000{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":"failed",
			"message": "order above 1000 is not allowed for COD",
		})
		return
	}
	err:=utils.CartToOrderItem( ucart,order.OrderID )
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
			})
	}
	if paymentMethod==models.Wallet{
		if err:=utils.WalletPayment(userid,order.OrderID);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
	}
	var cu models.CouponUsage
	if err := database.DB.Model(&cu).Where("coupon = ? AND user_id = ?", couponcode, userid).Update("usage_count", cu.UsageCOunt+1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	} else {
		couponusage := models.CouponUsage{
			Coupon:     couponcode,
			UserID:     userid,
			UsageCOunt: 1,
		}
		if err := database.DB.Create(&couponusage).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"message":       "Order created successfully",
		"data":          order,
		"coupon Status": message,
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
	if err := database.DB.Model(&models.Order{}).Where("user_id=?",userid).Order("order_id desc").Find(&ord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve orders",
		})
		return
	}
	var ordresponse []models.OrderResponse
	for _, ords := range ord {
		var ordItems []models.OrderItem
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
				Price:       prodct.Price,
				FinalAmount: prodct.OfferPrice,
				Quantity:    ordItem.Quantity,
				TotalPrice:  RoundDecimalValue(float64(ordItem.Quantity) * prodct.OfferPrice),
				OrderStatus: ordItem.OrderStatus,
			})
		}
		ordresponse = append(ordresponse, models.OrderResponse{
			OrderID:         ords.OrderID,
			OrderDate:       ords.OrderDate,
			SubtotalAmount:       ords.RawAmount,
			OfferTotal:      ords.OfferTotal,
			DiscountAmount:   RoundDecimalValue(ords.DiscountAmount),
			ShippingCharge: ords.DeliveryCharge,
			TotalPayable:     RoundDecimalValue(ords.FinalAmount),
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
	userid, _ := utils.GetUserIDByEmail(emailval)
	ordid := c.Query("order_id")
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
	var order models.Order
	if err:=database.DB.Where("order_id=?",orderID).First(&order).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order",
		})
		return
	}
	if order.OrderStatus==models.Cancelled{
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order is already cancelled"})
		return
	}
	if err := utils.CancelOrder(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	var user models.User
	if err:=database.DB.Where("id=?",userid).First(&user).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status": "failed",
			"message": err.Error(),	
			})
			return
	}
	cancelamount:=user.WalletAmount+order.FinalAmount
	if err:=database.DB.Model(&user).Where("id=?",userid).Update("wallet_amount",cancelamount).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"message": err.Error(),
			})
			return
	}
	walletTransaction:=models.UserWallet{
		UserID:userid,
		Amount:uint(cancelamount),
		OrderId: orderID,
		WalletPaymentId: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		TypeOfPayment: "incoming",
		TransactionTime: time.Now(),
		CurrentBalance: uint(user.WalletAmount),
		Reason: "cancel",
	}
	if err:=database.DB.Create(&walletTransaction).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order cancelled successfully"})
}
func CancelItemFromUserOrders(c *gin.Context) {
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
	userid, _ := utils.GetUserIDByEmail(emailval)
	ordid := c.Query("order_id")
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
	cancelprice,err,final := utils.CancelOrderItem(uint(parsedOrderID), uint(parsedProductID))
	if  err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	if cancelprice==models.PointThree{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":  "failed",
			"message": "the item is already cancelled",
		})
		return
	}
	if cancelprice==0{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":  "failed",
			"message": "the item is not delivered yet",
		})
		return
	}
	var coupon models.Coupon
	database.DB.Where("order_id=?",ordid).First(&coupon)
	if final!=0&&final<coupon.MinimumPurchase{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":  "failed",
			"message": "coupon cant applied to this order",
		})
		return
	}
	var user models.User
	if err:=database.DB.Where("id=?",userid).First(&user).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to find the user"})
		return
	}
	price:=user.WalletAmount+cancelprice
	if err:=database.DB.Model(&models.User{}).Where("id=?",userid).Update("wallet_amount",price).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	var orditems []models.OrderItem
	if err:=database.DB.Model(&orditems).Where("order_id=?",ordid).Find(&orditems).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	flag:=0
	for _,item:=range orditems{
		if item.OrderStatus==models.Cancelled{
			flag++
		}
	}
	if flag==len(orditems){
		database.DB.Model(&models.Order{}).Where("order_id=?",ordid).Update("order_status",models.Cancelled)
	}
	var order models.Order
	if err:=database.DB.Where("order_id=?",ordid).First(&order).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	if order.PaymentStatus==models.Success{
		walletTransaction:=models.UserWallet{
			UserID:userid,
			Amount:uint(cancelprice),
			OrderId: uint(parsedOrderID),
			WalletPaymentId: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
			TypeOfPayment: "incoming",
			TransactionTime: time.Now(),
			CurrentBalance: uint(user.WalletAmount),
			Reason: "cancel",
		}
		if err:=database.DB.Create(&walletTransaction).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}
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
			if err := database.DB.Model(&prodct).Where("product_id=?", orditem.ProductID).First(&prodct).Error; err != nil {
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
			OfferTotal:      order.OfferTotal,
			DiscountPrice:   order.DiscountAmount,
			FinalAmount:     order.FinalAmount,
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

	ordid := c.Query("order_id")
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
	var order models.Order
	if err:=database.DB.Where("order_id=?",orderID).First(&order).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order",
		})
		return
	}
	if order.OrderStatus==models.Cancelled{
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order is already cancelled"})
		return
	}

	if err := utils.CancelOrder(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order cancelled successfully"})
}

func CancelItemFromAdminOrders(c *gin.Context) {
	ordid := c.Query("order_id")
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
	 cancel,err,_ := utils.CancelOrderItem(uint(parsedOrderID), uint(parsedProductID))
	 if  err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	if cancel==models.PointThree{
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Order item is already cancelled"})
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

	// Retrieve order ID from query parameter
	ordid := c.Query("order_id")
	if ordid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Order ID is required"})
		return
	}

	// Convert order ID to uint
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

	// Retrieve order items from the database
	var orditems []models.OrderItem
	if err := database.DB.Where("order_id=?", orderID).Find(&orditems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order items",
		})
		return
	}

	// Check if order is already cancelled
	if ord.OrderStatus == models.Cancelled {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Order is already cancelled",
		})
		return
	}

	// Initialize transaction
	tx := database.DB.Begin()

	// Update the order and its items based on the current status
	var updatedStatus string
	switch ord.OrderStatus {
	case models.Pending:
		updatedStatus = models.Confirm
	case models.Confirm:
		updatedStatus = models.Shipped
	case models.Shipped:
		updatedStatus = models.Delivered
	}
	if updatedStatus != "" {
		if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", orderID).Update("order_status", updatedStatus).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotModified, gin.H{
				"status":  "failed",
				"message": "Unable to update order item status",
			})
			return
		}
		ord.OrderStatus = updatedStatus
		if updatedStatus == models.Delivered {
			ord.PaymentStatus = models.Success
		}
		if err := tx.Save(&ord).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to save updated order",
			})
			return
		}
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Transaction failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":                   "success",
			"message":                  "Order status updated successfully",
			"your order status is now": updatedStatus,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid order status transition",
		})
	}
}
func ReturnOrder(c *gin.Context){
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
	userid, _ := utils.GetUserIDByEmail(emailval)
	ordID := c.Query("order_id")
	productid:=c.Query("productid")
	parsedID, _ := strconv.ParseUint(ordID, 10, 32)
	parsedId, _ := strconv.ParseUint(productid, 10, 32)
	prodid:=uint(parsedId,)
	orderid:=uint(parsedID)
	var order models.Order
	if err:=database.DB.Where("order_id=?",ordID).First(&order).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order",
			})
			return
	}
	returnprice,err:=utils.ReturnOrderItem(orderid,prodid)
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to return order item",
			})
			return
	}
	if returnprice==models.PointThree{
		c.JSON(http.StatusOK, gin.H{
			"status":  "failed",
			"message": "You can't return this product",
			})
			return
	}
	var user models.User
	if err:=database.DB.Where("id=?",userid).First(&user).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user",
			})
			return
	}
	price:=user.WalletAmount+returnprice
	if err:=database.DB.Model(&user).Where("id=?",userid).Update("wallet_amount",price).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"status":  "failed",
			"message": "Failed to update wallet amount",
		})
		return
	}
	walletTransaction:=models.UserWallet{
		UserID:userid,
		Amount:uint(returnprice),
		OrderId: orderid,
		WalletPaymentId: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		TypeOfPayment: "incoming",
		TransactionTime: time.Now(),
		CurrentBalance: uint(user.WalletAmount),
		Reason: "cancel",
	}
	if err:=database.DB.Create(&walletTransaction).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":  "success",
		"message": "Order item returned successfully",
	})
}
