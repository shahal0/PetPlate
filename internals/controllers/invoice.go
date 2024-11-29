package controllers

import (
	"fmt"
	"petplate/internals/database"
	"petplate/internals/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)

func Invoice(c *gin.Context) {
	orderID := c.Query("order_id")
	orderDetails, err := fetchOrderDetailsFromDB(orderID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to fetch order details"})
		return
	}
	var order models.Order
	if err := database.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(500, gin.H{"error": "Unable to fetch order details for totals"})
		return
	}
	if order.PaymentStatus!=models.Success{
		c.JSON(500, gin.H{"error": "Payment status is not success"})
		return
	}
	// Fetching items
	var items []models.OrderItem
	if err := database.DB.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		c.JSON(500, gin.H{"error": "Unable to fetch order items"})
		return
	}

	// Invoice details
	invoiceNumber := fmt.Sprintf("INV-%s", orderID)
	date := time.Now().Format("02 Jan 2006")
	dueDate := time.Now().AddDate(0, 0, 15).Format("02 Jan 2006")

	// Seller Info
	sellerInfo := fmt.Sprintf("Axis Bank\nAccount Name: petplate\nAccount No.: 123-456-7890\nPay by: %v",orderDetails.OrderDate)

	// Buyer Info
	buyerInfo := fmt.Sprintf("%s\nPhone: %v\n%s %s\n%s, %s %s",
		orderDetails.CustomerName,
		orderDetails.CustomerAddress.PhoneNumber,
		orderDetails.CustomerAddress.StreetName,
		orderDetails.CustomerAddress.StreetNumber,
		orderDetails.CustomerAddress.City,
		orderDetails.CustomerAddress.PostalCode,
		orderDetails.CustomerAddress.State,
	)

	// Initialize PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 15, "Tax Invoice", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Invoice and Date
	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(130, 20)
	pdf.CellFormat(0, 6, fmt.Sprintf("Invoice No: %s", invoiceNumber), "", 1, "R", false, 0, "")
	pdf.SetXY(130, 26)
	pdf.CellFormat(0, 6, fmt.Sprintf("Date: %s", date), "", 1, "R", false, 0, "")
	pdf.SetXY(130, 32)
	pdf.CellFormat(0, 6, fmt.Sprintf("Due Date: %s", dueDate), "", 1, "R", false, 0, "")

	// Billing and Shipping
	pdf.Ln(15)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 10, "Billed To:", "", 1, "", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, sellerInfo, "", "L", false)
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 10, "Shipped To:", "", 1, "", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, buyerInfo, "", "L", false)

	// Items Table
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(80, 10, "Item", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Total", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 12)
	var subtotal float64
	for _, item := range items {
		var product models.Product
		if err := database.DB.Where("product_id = ?", item.ProductID).First(&product).Error; err != nil {
			continue
		}
		total := float64(item.Quantity) * product.Price
		subtotal+=total
		pdf.CellFormat(80, 10, product.Name, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", product.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", total), "1", 1, "R", false, 0, "")
	}

	// Additional Charges and Final Total

	totalOfferAmount := order.RawAmount - order.OfferTotal
	totalDiscountAmount := order.DiscountAmount
	totalDeliveryCharge := order.DeliveryCharge
	finalGrandTotal :=  order.FinalAmount

	// Totals Section
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(150, 10, "Subtotal", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", subtotal), "1", 1, "R", false, 0, "")
	pdf.CellFormat(150, 10, "Total Offer Amount", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", totalOfferAmount), "1", 1, "R", false, 0,"")
	pdf.CellFormat(150, 10, "Total Discount Amount", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", totalDiscountAmount), "1", 1, "R", false, 0,"")
	pdf.CellFormat(150, 10, "Total Delivery Charge", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", totalDeliveryCharge), "1", 1, "R", false, 0,"")
	pdf.CellFormat(150, 10, "Grand Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", finalGrandTotal), "1", 1, "R", false, 0,"")

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=invoice_%s.pdf", invoiceNumber))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
}
func fetchOrderDetailsFromDB(orderID string) (models.OrderDetails, error) {
	var order models.Order
	if err := database.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return models.OrderDetails{}, err
	}

	var user models.User
	if err := database.DB.Where("id = ?", order.UserID).First(&user).Error; err != nil {
		return models.OrderDetails{}, err
	}

	var orderItems []models.OrderItem
	if err := database.DB.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		return models.OrderDetails{}, err
	}

	// Populate Invoice Items
	var items []models.InvoiceItem
	for _, orderItem := range orderItems {
		var product models.Product
		if err := database.DB.Where("product_id = ?", orderItem.ProductID).First(&product).Error; err != nil {
			return models.OrderDetails{}, err
		}
		items = append(items, models.InvoiceItem{
			Name:     product.Name,
			Quantity: orderItem.Quantity,
			Price:    product.Price,
		})
	}

	// Assemble Order Details
	orderDetails := models.OrderDetails{
		CustomerName:    user.Name,
		CustomerAddress: order.ShippingAddress,
		CustomerCity:    order.ShippingAddress.City,
		OrderDate:       order.OrderDate,
		Items:           items,
	}

	return orderDetails, nil
}
