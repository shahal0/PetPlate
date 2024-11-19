// package controllers

package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)
var count uint
func init() {
	err := godotenv.Load() // Load .env file
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func RenderRayzorPay(c *gin.Context) {
	orderId := c.Query("orderid")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	c.HTML(http.StatusOK, "payment.html", gin.H{
		"orderID": orderId, // Replace with any necessary data
	})
}

func CreateOrder(c *gin.Context) {
	log.Println("Entered CreateOrder function")
	orderid := c.Query("orderid")
	if orderid == "" {
		log.Println("Order ID is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	// var orderrr models.Order
	// if err:=database.DB.Where("order_id=?",orderid).First(&orderrr).Error;err!=nil{
	// 	log.Println(err)
	// 	c.JSON(http.StatusBadRequest,gin.H{
	// 		"error": "Order not found",
	// 	})
	// }
	// if orderrr.PaymentMethod!="UPI"{
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Payment method is not UPI"})
	// 	log.Println("payment method is not UPI")
	// }

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	var orderr models.Order
	if err := database.DB.Where("order_id = ?", orderid).Find(&orderr).Error; err != nil {
		log.Println("Order not found in database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order not found"})
		return
	}
	if orderr.PaymentStatus==models.Success||orderr.PaymentStatus==models.OrderPaid{
		c.JSON(http.StatusOK, gin.H{"error": "Order already paid"})
		return
	}
	amount := int(orderr.FinalAmount * 100)
	log.Println("Creating Razorpay order")
	data := map[string]interface{}{
		"amount":          amount, // Amount in paise
		"currency":        "INR",
		"payment_capture": 1,
	}

	order, err := client.Order.Create(data, nil)
	if err != nil {
		log.Println("Error creating order with Razorpay:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create order"})
		return
	}

	log.Println("Order created successfully:", order)
	c.JSON(http.StatusOK, order)
}

func VerifyPayment(c *gin.Context) {
	log.Println("Entered VerifyPayment function")

	// Get the Order ID from the URL parameter
	orderId := c.Param("orderid")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	// Bind JSON data to retrieve Razorpay payment details
	var PaymentInfo struct {
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}
	if err := c.BindJSON(&PaymentInfo); err != nil {
		log.Println("BindJSON error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Retrieve Razorpay key secret from environment variables
	razorpayKeySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	if razorpayKeySecret == "" {
		log.Println("RAZORPAY_KEY_SECRET not found in environment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}

	// Generate the expected signature
	data := PaymentInfo.RazorpayOrderID + "|" + PaymentInfo.RazorpayPaymentID
	expectedSignature := generateRazorpaySignature(data, razorpayKeySecret)
	log.Println("Expected Signature:", expectedSignature)


	// Start a transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		log.Println("Transaction start error:", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to start transaction"})
		return
	}

	// Check if a payment record already exists for this order ID
	var existingPayment models.Payment
	if err := tx.Where("order_id = ?", orderId).First(&existingPayment).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment already exists for this order"})
		tx.Rollback()
		return
	}

	// Create a new payment record
	payment := models.Payment{
		OrderID:            orderId,
		RayzorpayOrderID:   PaymentInfo.RazorpayOrderID,
		RayzorPayPaymentID: PaymentInfo.RazorpayPaymentID,
		RayzorPaySignature: PaymentInfo.RazorpaySignature,
		PaymentGateway:     models.RayzorPay,
		PaymentStatus:      models.PaymentPending,
	}

	// Save the payment record
	if err := tx.Create(&payment).Error; err != nil {
		log.Println("Error creating payment record:", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
		return
	}

	// Update the order status to 'Paid'
	if err := tx.Model(&models.Order{}).Where("order_id = ?", orderId).Update("payment_status", models.OrderPaid).Error; err != nil {
		log.Println("Error updating order status:", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Update product stock based on order items
	var orderItems []models.OrderItem
	if err := tx.Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		log.Println("Error fetching order items:", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
		return
	}
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).Where("product_id = ?", item.ProductID).Update("max_stock", gorm.Expr("max_stock - ?", item.Quantity)).Error; err != nil {
			log.Println("Error updating product stock for product_id:", item.ProductID, "Error:", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	// Update the payment and order statuses to success after transaction commit
	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	// Finalize payment status
	if err := database.DB.Model(&payment).Where("payment_id = ?", payment.PaymentID).Update("payment_status", models.Success).Error; err != nil {
		log.Println("Error updating payment status:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}
	if err := database.DB.Model(&models.Order{}).Where("order_id = ?", orderId).Update("payment_status", models.Success).Error; err != nil {
		log.Println("Error updating order payment status:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order payment status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Payment verified and recorded successfully"})
}

// Helper function to generate HMAC SHA256 signature
func generateRazorpaySignature(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func FailedHandling(c *gin.Context) {
	orderid := c.Query("orderid")
	var order models.Order
	if orderid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order id"})
		return
	}
	if count>=3{
		database.DB.Model(&order).Where("order_id=?", orderid).Update("payment_status", models.PaymentFailed)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many failed attempts."})
		return
	}
	if err := database.DB.Model(&order).Where("order_id=?", orderid).Update("payment_status", models.PaymentFailed).Error; err != nil {
		log.Println(err)
		return
	}
	count++

	var payment models.Payment
	if err := database.DB.Model(&payment).Where("rayzorpay_order_id=?", payment.RayzorpayOrderID).Update("payment_status", models.PaymentFailed).Error; err != nil {
		log.Println(err)
	}
}
