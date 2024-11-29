// package controllers

package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	orderid := c.Query("orderid")
	if orderid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	var orderr models.Order
	if err := database.DB.Where("order_id = ?", orderid).Find(&orderr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order not found"})
		return
	}
	if orderr.PaymentStatus==models.Success||orderr.PaymentStatus==models.OrderPaid{
		c.JSON(http.StatusOK, gin.H{"error": "Order already paid"})
		return
	}
	amount := int(orderr.FinalAmount * 100)
	data := map[string]interface{}{
		"amount":          amount, // Amount in paise
		"currency":        "INR",
		"payment_capture": 1,
	}

	order, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func VerifyPayment(c *gin.Context) {

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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
        return
    }


    // Start a transaction
    tx := database.DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    if tx.Error != nil {
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
        RayzorPaySignature: PaymentInfo.RazorpaySignature, // May be empty
        PaymentGateway:     models.RayzorPay,
        PaymentStatus:      models.PaymentPending,
    }

    // Save the payment record
    if err := tx.Create(&payment).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
        return
    }

    // Update the order status to 'Paid'
    if err := tx.Model(&models.Order{}).Where("order_id = ?", orderId).Update("payment_status", models.Success).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
        return
    }

    // Update product stock based on order items
    var orderItems []models.OrderItem
    if err := tx.Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
        return
    }
    for _, item := range orderItems {
        if err := tx.Model(&models.Product{}).Where("product_id = ?", item.ProductID).Update("max_stock", gorm.Expr("max_stock - ?", item.Quantity)).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
            return
        }
    }

    // Commit the transaction
    if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
        return
    }

    // Finalize payment status
    if err := database.DB.Model(&payment).Where("payment_id = ?", payment.PaymentID).Update("payment_status", models.Success).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
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
		count++
		return
	}

	var payment models.Payment
	if err := database.DB.Model(&payment).Where("rayzorpay_order_id=?", payment.RayzorpayOrderID).Update("payment_status", models.PaymentFailed).Error; err != nil {
	}
	count=0
}
