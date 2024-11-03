package utils

import (
	"fmt"
	"net/http"
	"petplate/internals/database"
	"petplate/internals/models"

	"github.com/gin-gonic/gin"
)

// GetUserIDByEmail retrieves the user ID based on the provided email address.
func GetUserIDByEmail(email string) (uint, bool) {
	var userID uint
	if err := database.DB.Model(&models.User{}).Where("email = ?", email).Pluck("id", &userID).Error; err != nil {
		return 0, false
	}
	return userID, true
}
func Itercart(ucart []models.Cart)float64{
	//var orderitems[]models.OrderItem
	var TotalAmount float64
	for _,cartitem:=range ucart{
		var prod  models.Product
		if err:=database.DB.Where("id=?",cartitem.ProductID).First(&prod).Error;err!=nil{
			// c.JSON(http.StatusInternalServerError,gin.H{
			// 	"status":"failed",
			// 	"message":err.Error(),
			// 	})
				// return
		}
		amount:=prod.Price*float64(cartitem.Quantity)
		
		// orditems:=models.OrderItem{
		// 	OrderID: ,
		// 	ProductID: cartitem.ProductID,
		// 	Quantity: cartitem.Quantity,
		// 	Price: amount,
		// 	OrderStatus: models.Pending ,
		// }
		// if err:=database.DB.Create(&orditems).Error;err!=nil{
		// 	c.JSON(http.StatusBadRequest,gin.H{
		// 		"status":  "failed",
		// 		"message": err.Error(),
		// 		})
		// 		return
		// }
		TotalAmount+=amount
		//orderitems=append(orderitems,orditems)
			
	}
	return TotalAmount
}
func CartToOrderItem(c gin.Context,ucart []models.Cart,orderid uint){
	for _,cartitem:=range ucart{
		var prod  models.Product
		if err:=database.DB.Where("id=?",cartitem.ProductID).First(&prod).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"status":"failed",
				"message":err.Error(),
				})
				return 
		}
		amount:=prod.Price*float64(cartitem.Quantity)
		
		orditems:=models.OrderItem{
			OrderID: orderid,
			ProductID: cartitem.ProductID,
			Quantity: cartitem.Quantity,
			Price: amount,
			OrderStatus: models.Pending ,
		}
		if err:=database.DB.Create(&orditems).Error;err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"status":  "failed",
				"message": err.Error(),
				})
				return
		}
		prod.MaxStock-=cartitem.Quantity	
	}
	return 
}
func CancelOrder(orderID uint) error {
	var order models.Order
	if err := database.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("failed to retrieve order: %w", err)
	}

	order.OrderStatus = models.Cancelled
	if err := database.DB.Save(&order).Error; err != nil {
		return fmt.Errorf("unable to save changes: %w", err)
	}

	return nil
}
func CancelOrderItem(orderID, productID uint) error {
	// Retrieve the order
	var order models.Order
	if err := database.DB.First(&order, "order_id = ?", orderID).Error; err != nil {
		return fmt.Errorf("failed to retrieve order: %w", err)
	}

	// Retrieve the order item
	var ordItem models.OrderItem
	if err := database.DB.First(&ordItem, "order_id = ? AND product_id = ?", orderID, productID).Error; err != nil {
		return fmt.Errorf("failed to retrieve order item: %w", err)
	}

	// Update the order item status to Cancelled
	if err := database.DB.Model(&ordItem).Update("order_status", models.Cancelled).Error; err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	// Recalculate the total for remaining non-cancelled items
	var remainingItems []models.OrderItem
	if err := database.DB.Where("order_id = ? AND order_status != ?", orderID, models.Cancelled).Find(&remainingItems).Error; err != nil {
		return fmt.Errorf("failed to retrieve remaining order items: %w", err)
	}

	// Calculate the new total
	newTotal := 0.0
	for _, item := range remainingItems {
		newTotal += item.Price
	}

	// Update the order total
	if err := database.DB.Model(&models.Order{}).Where("order_id = ?", orderID).Update("total", newTotal).Error; err != nil {
		return fmt.Errorf("failed to update order total: %w", err)
	}

	return nil
}

