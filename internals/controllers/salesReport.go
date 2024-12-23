package controllers

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"os"
	"petplate/internals/database"
	"petplate/internals/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/wcharczuk/go-chart"
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

	// Parse the input dates
	startDate, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing start date: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing end date: %v", err)
	}

	// Adjust end date for full-day precision
	startDate = startDate.UTC()
	endDate = endDate.Add(24*time.Hour - time.Nanosecond).UTC()

	// Fetch orders within the date range and payment status
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

	// If no orders exist, return early
	if len(orders) == 0 {
		return models.OrderCount{TotalOrder: 0}, accountInfo, nil
	}

	// Fetch status counts for the orders
	var statusCounts []struct {
		OrderStatus string
		Count       int64
	}
	if err := database.DB.Raw(`
		SELECT order_status, COUNT(*) as count 
		FROM orders 
		WHERE order_date BETWEEN ? AND ? AND payment_status = ?
		GROUP BY order_status
	`, startDate, endDate, paymentStatus).Scan(&statusCounts).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error counting order items by status: %v", err)
	}

	// Map status counts
	orderStatusCounts := make(map[string]int64)
	var totalCount int64
	for _, sc := range statusCounts {
		orderStatusCounts[sc.OrderStatus] = sc.Count
		totalCount += sc.Count
	}

	// Safely map and return the results
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
    pdf.SetFont("Arial", "B", 20)
    pdf.CellFormat(0, 10, "Sales Report", "", 1, "C", false, 0, "")
    pdf.Ln(10)

    
    pdf.SetFont("Arial", "B", 12)
    pdf.CellFormat(40, 10, "Report Duration:", "", 1, "L", false, 0, "")
    pdf.SetFont("Arial", "", 12)
    pdf.CellFormat(40, 10, "Start Date: "+startDate, "", 1, "L", false, 0, "")
    pdf.CellFormat(40, 10, "End Date: "+endDate, "", 1, "L", false, 0, "")
    pdf.CellFormat(40, 10, "Payment Status: "+paymentStatus, "", 1, "L", false, 0, "")
    pdf.Ln(12)


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

    // Order History Table
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Order History Details")
    pdf.Ln(8)

    pdf.SetFont("Arial", "B", 12)
    pdf.CellFormat(60, 10, "Status", "1", 0, "C", false, 0, "")
    pdf.CellFormat(60, 10, "Count", "1", 0, "C", false, 0, "")
    pdf.Ln(-1)

    pdf.SetFont("Arial", "", 12)
    orderHistory := map[string]int{
        "Pending":   int(orderCount.TotalPending),
        "Confirmed": int(orderCount.TotalConfirmed),
        "Shipped":   int(orderCount.TotalShipped),
        "Delivered": int(orderCount.TotalDelivered),
        "Cancelled": int(orderCount.TotalCancelled),
        "Returned":  int(orderCount.TotalReturned),
    }

    for status, count := range orderHistory {
        pdf.CellFormat(60, 10, status, "1", 0, "L", false, 0, "")
        pdf.CellFormat(60, 10, strconv.Itoa(count), "1", 0, "R", false, 0, "")
        pdf.Ln(-1)
    }
    pdf.Ln(10)


    // Prepare Chart Data
    chartData := []chart.Value{
        {Value: float64(orderCount.TotalPending), Label: "Pending"},
        {Value: float64(orderCount.TotalConfirmed), Label: "Confirmed"},
        {Value: float64(orderCount.TotalShipped), Label: "Shipped"},
        {Value: float64(orderCount.TotalDelivered), Label: "Delivered"},
        {Value: float64(orderCount.TotalCancelled), Label: "Cancelled"},
        {Value: float64(orderCount.TotalReturned), Label: "Returned"},
    }

    // Validate Chart Data
    validData := false
    for _, data := range chartData {
        if data.Value > 0 {
            validData = true
            break
        }
    }

    if validData {
        // Create bar chart
        barChart := chart.BarChart{
            Width:  500,
            Height: 300,
            Bars:   chartData,
            XAxis: chart.Style{
                Show: true,
            },
            YAxis: chart.YAxis{
                Style: chart.Style{
                    Show: true,
                },
                Range: &chart.ContinuousRange{
                    Min: 0,
                    Max: float64(orderCount.TotalOrder),
                },
            },
        }

        var chartBuffer bytes.Buffer
        err := barChart.Render(chart.PNG, &chartBuffer)
        if err != nil {
            return nil, fmt.Errorf("failed to generate bar chart: %v", err)
        }

        chartFileName := "temp_chart.png"
        err = os.WriteFile(chartFileName, chartBuffer.Bytes(), 0644)
        if err != nil {
            return nil, fmt.Errorf("failed to save chart image: %v", err)
        }
        defer os.Remove(chartFileName)


		tableWidth := 90 + 60
	pageWidth, pageHeight := pdf.GetPageSize()

	
	chartWidth := float64(tableWidth) / 1.4
	chartHeight := chartWidth / 1.5 
	remainingHeight := pageHeight - pdf.GetY() - 10 
	if chartHeight > remainingHeight {
		pdf.AddPage() 
		pdf.Ln(5)     
	}
		// Add heading for the bar chart
	pdf.SetFont("Arial", "B", 14) 
	pdf.CellFormat(0, 10, "Order Status Distribution", "", 1, "C", false, 0, "")
	pdf.Ln(5) 


	// Center the chart horizontally on the page
	chartX := (pageWidth - chartWidth) / 2
	chartY := pdf.GetY() + 2

	// Place the chart
	pdf.ImageOptions(
		chartFileName,
		chartX,
		chartY,
		chartWidth,
		chartHeight,
		false,
		gofpdf.ImageOptions{ImageType: "PNG"},
		0,
		"",
	)
	pdf.SetY(chartY + chartHeight + 2)


    } else {
        pdf.SetFont("Arial", "I", 12)
        pdf.CellFormat(0, 10, "No data available for chart representation.", "", 1, "C", false, 0, "")
        pdf.Ln(10)
    }

    // Generate the PDF output
    var buf bytes.Buffer
    err := pdf.Output(&buf)
    if err != nil {
        return nil, fmt.Errorf("error generating PDF: %v", err)
    }

    return buf.Bytes(), nil
}







