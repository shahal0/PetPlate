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
)
func init() {
    err := godotenv.Load() // Load .env file
    if err != nil {
        log.Fatalf("Error loading .env file")
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

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	var orderr models.Order
	if err := database.DB.Where("order_id = ?", orderid).Find(&orderr).Error; err != nil {
		log.Println("Order not found in database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order not found"})
		return
	}
	amount:=int(orderr.FinalAmount*100)
	log.Println("Creating Razorpay order")
	data := map[string]interface{}{
		"amount":           amount,// Amount in paise
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
    log.Println("entered verify")
    orderId := c.Param("orderid")
    if orderId == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
        return
    }

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

    log.Println("Payment Info:", PaymentInfo) 
    if PaymentInfo.RazorpaySignature == "" {
        log.Println("Received empty RazorpaySignature")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Signature is required"})
        return
    }

    expectedSignature := generateRazorpaySignature(
        PaymentInfo.RazorpayOrderID+"|"+PaymentInfo.RazorpayPaymentID,
        os.Getenv("RAZORPAY_KEY_SECRET"),
    )
    log.Println("Generated Signature:", expectedSignature)
    log.Println("Received Signature:", PaymentInfo.RazorpaySignature)

    if PaymentInfo.RazorpaySignature == expectedSignature {
        c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Payment verified successfully"})
    } else {
        log.Println("Signature mismatch")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Signature mismatch"})
        return
    }
}

func generateRazorpaySignature(data, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
func FailedHandling(c *gin.Context) {
	orderid := c.Query("orderid")
	if orderid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order id"})
		return
	}

	var order models.Order
	if err := database.DB.Model(&order).Where("order_id=?", orderid).Update(order.PaymentStatus, models.PaymentFailed).Error; err != nil {
		log.Println(err)
		return

	}
	var payment models.Payment
	if err:=database.DB.Model(&payment).Where("rayzorpay_order_id=?",payment.RayzorpayOrderID).Update(payment.PaymentStatus,models.PaymentFailed).Error;err!=nil{
		log.Println(err)
	}
	
}
