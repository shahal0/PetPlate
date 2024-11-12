package controllers

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)
func RoundDecimalValue(value float64) float64 {
	multiplier := math.Pow(10, 2)
	return math.Round(value*multiplier)/multiplier
}
func SalesReport(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit := c.Query("limit")
	pStatus := c.Query("payment_status")

	if startDate == "" && endDate == "" && limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "please provide start date and end date, or specify the limit as day, week, month, year",
		})
		return
	}

	// Handle `limit` as a time range (day, week, month, year)
	if limit != "" {
		switch limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
			startDate = startOfWeek.Format("2006-01-02")
			endDate = startOfWeek.AddDate(0, 0, 7).Format("2006-01-02")
		case "month":
			today := time.Now()
			startDate = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location()).Format("2006-01-02")
			endDate = time.Date(today.Year(), today.Month()+1, 0, 0, 0, 0, 0, today.Location()).Format("2006-01-02")
		case "year":
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}
	}

	// Call TotalOrders to retrieve results within date range
	result, amount, err := TotalOrders(startDate, endDate, pStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
	})
}

func TotalOrders(fromDate string, toDate string, paymentStatus string) (models.OrderCount, models.AmountInformation, error) {
	var orders []models.Order

	// Parse and adjust date range
	startDate, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing start date: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing end date: %v", err)
	}

	// Adjust to full-day precision for the end date
	startDate = startDate.UTC()
	endDate = endDate.Add(24*time.Hour - time.Nanosecond).UTC()

	// Query orders within the date range and payment status
	if err := database.DB.
		Where("order_date BETWEEN ? AND ? AND payment_status = ?", startDate, endDate, paymentStatus).
		Find(&orders).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error fetching orders: %v", err)
	}

	// Aggregate financial details
	var accountInfo models.AmountInformation
	for _, order := range orders {
		accountInfo.TotalAmountBeforeDeduction += order.RawAmount
		accountInfo.TotalCouponDeduction += order.DiscountAmount
		accountInfo.TotalProuctOfferDeduction += order.OfferTotal
		accountInfo.TotalAmountAfterDeduction += order.FinalAmount
	}

	// Extract order IDs for status count query
	// var orderIDs []int64
	// for _, order := range orders {
	// 	orderIDs = append(orderIDs, int64(order.OrderID))
	// }

	// Check if orderIDs is empty
	if len(orders) == 0 {
		return models.OrderCount{TotalOrder: 0}, accountInfo, nil
	}

	// Query order statuses and counts
	var statusCounts []struct {
		OrderStatus string
		Count       int64
	}
	if err := database.DB.Raw(`
		SELECT order_status, COUNT(*) as count 
		FROM "order_items" 
		GROUP BY "order_status"
	`).Scan(&statusCounts).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error counting order items by status: %v", err)
	}
	log.Println(statusCounts)

	// Map status counts
	orderStatusCounts := make(map[string]int64)
	var totalCount int64
	for _, sc := range statusCounts {
		orderStatusCounts[sc.OrderStatus] = sc.Count
		totalCount += sc.Count
	}

	// Safely access counts for each status
	return models.OrderCount{
		TotalOrder:     uint(totalCount),
		TotalPending:   uint(orderStatusCounts[models.Pending]),
		TotalConfirmed: uint(orderStatusCounts[models.Confirm]),
		TotalShipped:   uint(orderStatusCounts[models.Shipped]),
		TotalDelivered: uint(orderStatusCounts[models.Delivered]),
		TotalCancelled: uint(orderStatusCounts[models.Cancelled]),
		TotalReturned:  uint(orderStatusCounts[models.Return]),
	}, accountInfo, nil
}
func DownloadSalesReportPDF(c *gin.Context) {
	_, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Unauthorized or invalid token",
		})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	paymentStatus := c.Query("payment_status")
	limit := c.Query("limit")

	if limit == "" && (startDate == "" || endDate == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide either limit or start_date and end_date"})
		return
	}
	if limit != "" {
		switch limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			startDate = time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
			endDate = time.Now().AddDate(0, 0, 7-int(time.Now().Weekday())).Format("2006-01-02")
		case "month":
			startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "year":
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}
	}

	orderCount, amountInfo, err := TotalOrders(startDate, endDate, paymentStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	pdfBytes, err := GenerateSalesReportPDF(orderCount, amountInfo, startDate, endDate, paymentStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func GenerateSalesReportPDF(orderCount models.OrderCount, amountInfo models.AmountInformation, startDate string, endDate string, paymentStatus string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Centered Report Title
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 10, "Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Left-Aligned Date & Filter Information (Report Duration)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "Report Duration:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(40, 10, "Start Date: "+startDate, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 10, "End Date: "+endDate, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 10, "Payment Status: "+paymentStatus, "", 1, "L", false, 0, "")
	pdf.Ln(12)

	// Section: Summary Information Table
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Summary Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(90, 10, "Description", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, "Amount", "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 12)
	summaryData := map[string]string{
		"Total Orders":                   strconv.Itoa(int(orderCount.TotalOrder)),
		"Total Amount Before Deduction":  fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction),
		"Total Coupon Deduction":         fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction),
		"Total Product Offer Deduction":  fmt.Sprintf("%.2f", amountInfo.TotalProuctOfferDeduction),
		"Total Amount After Deduction":   fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction),
	}

	for desc, amount := range summaryData {
		pdf.CellFormat(90, 10, desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 10, amount, "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}
	pdf.Ln(10)

	// Section: Order Status Summary Table
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Order Status Summary")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(90, 10, "Order Status", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, "Total Count", "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 12)
	orderStatuses := map[string]uint{
		"Pending Orders":   orderCount.TotalPending,
		"Confirmed Orders": orderCount.TotalConfirmed,
		"Shipped Orders":   orderCount.TotalShipped,
		"Delivered Orders": orderCount.TotalDelivered,
		"Cancelled Orders": orderCount.TotalCancelled,
		"Returned Orders":  orderCount.TotalReturned,
	}

	for status, count := range orderStatuses {
		pdf.CellFormat(90, 10, status, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 10, strconv.Itoa(int(count)), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Generate PDF output
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
