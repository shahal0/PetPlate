package utils

import (
	"errors"
	"fmt"
	"petplate/internals/database"
	"petplate/internals/models"
	"time"

	
	
)

// GetUserIDByEmail retrieves the user ID based on the provided email address.
func GetUserIDByEmail(email string) (uint, bool) {
	var userID uint
	if err := database.DB.Model(&models.User{}).Where("email = ?", email).Pluck("id", &userID).Error; err != nil {
		return 0, false
	}
	return userID, true
}
func Itercart(ucart []models.Cart)(float64,float64){
	//var orderitems[]models.OrderItem
	var TotalAmount float64
	var RawAmount float64
	for _,cartitem:=range ucart{
		var prod  models.Product
		if err:=database.DB.Where("product_id=?",cartitem.ProductID).First(&prod).Error;err!=nil{
			// c.JSON(http.StatusInternalServerError,gin.H{
			// 	"status":"failed",
			// 	"message":err.Error(),
			// 	})
				 return 0,0
		}
		

		amount:=prod.OfferPrice*float64(cartitem.Quantity)
		ramount:=prod.Price*float64(cartitem.Quantity)
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
		var category models.Category
		database.DB.Model(&category).Where("id=?",prod.CategoryID).First(&category)
		TotalAmount+=amount
		RawAmount+=ramount
		amount-=prod.Price*(category.CategoryOffer/100)	

		//orderitemsppend(orderitems,orditems)
			
	}
	return TotalAmount,RawAmount
}
func CartToOrderItem(ucart []models.Cart,orderid uint)error{
	for _,cartitem:=range ucart{
		var prod  models.Product
		if err:=database.DB.Where("product_id=?",cartitem.ProductID).First(&prod).Error;err!=nil{
			// c.JSON(http.StatusInternalServerError,gin.H{
			// 	"status":"failed",
			// 	"message":err.Error(),
			// 	})
				return err
		}
		amount:=prod.OfferPrice*float64(cartitem.Quantity)
		
		orditems:=models.OrderItem{
			OrderID: orderid,
			ProductID: cartitem.ProductID,
			Quantity: cartitem.Quantity,
			Price: amount,
			OrderStatus: models.Pending ,
		}
		if err:=database.DB.Create(&orditems).Error;err!=nil{
			// c.JSON(http.StatusBadRequest,gin.H{
			// 	"status":  "failed",
			// 	"message": err.Error(),
			// 	})
				return err
		}
		prod.MaxStock-=cartitem.Quantity	
	}
	return nil
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
	if err:=database.DB.Model(&models.OrderItem{}).Where("order_id=?",orderID).Update("order_status",models.Cancelled).Error;err!=nil{
		return fmt.Errorf("failed to update order items: %w", err)
	}

	return nil
}
func CancelOrderItem(orderID, productID uint) (float64,error,float64) {
	// Retrieve the order
	var order models.Order
	if err := database.DB.First(&order, "order_id = ?", orderID).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve order: %w", err),0
	}

	// Retrieve the order item
	var ordItem models.OrderItem
	var cancelprice float64
	if err := database.DB.First(&ordItem, "order_id = ? AND product_id = ?", orderID, productID).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve order item: %w", err),0
	}
	if ordItem.OrderStatus==models.Cancelled{
		return models.PointThree,nil,0
	}

	// Update the order item status to Cancelled
	if err := database.DB.Model(&ordItem).Update("order_status", models.Cancelled).Error; err != nil {
		return 0,fmt.Errorf("failed to update order item status: %w", err),0
	}
	var coup models.Coupon
	if order.CouponCode!=""{
		if err:=database.DB.Model(&coup).Where("code=?",order.CouponCode).First(&coup).Error;err!=nil{
			return 0,fmt.Errorf("failed to retrieve coupon items: %w", err),0
		}
		cancelprice=ordItem.Price-(ordItem.Price*(coup.DiscountPercentage/100))
	}
	cancelprice=ordItem.Price
	// Recalculate the total for remaining non-cancelled items
	var remainingItems []models.OrderItem
	if err := database.DB.Where("order_id = ? AND order_status != ?", orderID, models.Cancelled).Find(&remainingItems).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve remaining order items: %w", err),0
	}

	// Calculate the new total
	newTotal := 0.0
	rawtotal:=0.0
	for _, item := range remainingItems {
		newTotal += item.Price
		var product models.Product
		database.DB.Model(&models.Product{}).Where("product_id=?",item.ProductID).First(&product)
		rawtotal+=product.Price*float64(item.Quantity)
	}
	

	// Update the order total
	if err := database.DB.Model(&models.Order{}).Where("order_id = ?", orderID).Update("total", newTotal).Error; err != nil {
		return 0,fmt.Errorf("failed to update order total: %w", err),0
	}
	price:=order.FinalAmount
	if err:=database.DB.Model(&order).Where("order_id=?",orderID).Update("raw_amount",rawtotal).Error;err!=nil{
		return 0,fmt.Errorf("failed to update order total: %w", err),0
	}
	// balance:=order.FinalAmount-cancelprice
	// if err:=database.DB.Model(&order).Where("order_id=?",orderID).Update("final_amount",balance).Error;err!=nil{
	// 	return 0,fmt.Errorf("failed to retrieve remaining order items: %w", err)
	// }
	newfinal:=order.FinalAmount-cancelprice
	if err:=database.DB.Model(&order).Where("order_id=?",orderID).Update("final_amount",newfinal).Error;err!=nil{
		return 0,fmt.Errorf("failed to update order total: %w", err),0
	}

	return cancelprice,nil,price
}
func FetchOrderDetails(orderid string)(models.Order,error){
	var order  models.Order

	database.DB.Model(&order).Where("order_id=?",orderid).First(&order)
	return order,nil
		
}
func Distercart(ucart []models.Cart,coup models.Coupon)(float64,float64){
	//var orderitems[]models.OrderItem
	var TotalAmount,amount float64
	// var RawAmount float64
	var disamount float64
	for _,cartitem:=range ucart{
		var prod  models.Product
		if err:=database.DB.Where("product_id=?",cartitem.ProductID).First(&prod).Error;err!=nil{
			// c.JSON(http.StatusInternalServerError,gin.H{
			// 	"status":"failed",
			// 	"message":err.Error(),
			// 	})
				// return
		}
		

		amount+=prod.OfferPrice*float64(cartitem.Quantity)
		// ramount:=prod.Price*float64(cartitem.Quantity)
		discount:=amount*(coup.DiscountPercentage/100)
		// RawAmount+=ramount
		disamount+=discount
			
	}
	if disamount>coup.MaximumDiscountAmount{
		disamount=coup.MaximumDiscountAmount
	}
	TotalAmount+=amount-disamount
	return TotalAmount,disamount
}
func WalletPayment(userid,orderid uint)error{
	var user models.User
	if err:=database.DB.Where("id=?",userid).First(&user).Error;err!=nil{
		return fmt.Errorf(err.Error())
	}
	var order models.Order
	if err:=database.DB.Where("order_id=?",orderid).First(&order).Error;err!=nil{
		return fmt.Errorf(err.Error())
	}
	if order.FinalAmount>user.WalletAmount{
		return errors.New("insufficient wallet amount")
	}
	updatedval:=user.WalletAmount-order.FinalAmount
	if err:=database.DB.Model(&user).Update("wallet_amount",updatedval).Error;err!=nil{
		return err
	}
	if err:=database.DB.Model(&order).Update("payment_status",models.Success).Error;err!=nil{
		return err
	}
	var orditems []models.OrderItem
	if err:=database.DB.Model(&models.OrderItem{}).Where("order_id=?",orderid).Find(&orditems).Error;err!=nil{
		return err
	}
	for _,item:=range orditems{
		var product models.Product
		if err:=database.DB.Model(&product).Where("product_id=?",item.ProductID).First(&product).Error;err!=nil{
			return err
		}
		if err:=database.DB.Model(&product).Update("max_stock",product.MaxStock-item.Quantity).Error;err!=nil{
			return err
		}
	}
	walletTransaction:=models.UserWallet{
		UserID:userid,
		Amount:uint(order.FinalAmount),
		OrderId: orderid,
		WalletPaymentId: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		TypeOfPayment: "outgoing",
		TransactionTime: time.Now(),
		CurrentBalance: uint(user.WalletAmount),
		Reason: "purchase",
	}
	if err:=database.DB.Create(&walletTransaction).Error;err!=nil{
		//c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
		return err
	}
	
	return nil
}
func ReturnOrderItem(orderID, productID uint) (float64,error) {
	// Retrieve the order
	var order models.Order
	if err := database.DB.First(&order, "order_id = ?", orderID).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve order: %w", err)
	}
	var ordItem models.OrderItem
	var cancelprice float64
	if err := database.DB.First(&ordItem, "order_id = ? AND product_id = ?", orderID, productID).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve order item: %w", err)
	}
	if ordItem.OrderStatus==models.Cancelled{
		return models.PointThree,nil
	}
	if ordItem.OrderStatus!=models.Delivered{
		return models.PointThree,nil
	}

	// Update the order item status to Cancelled
	if err := database.DB.Model(&ordItem).Update("order_status", models.Return).Error; err != nil {
		return 0,fmt.Errorf("failed to update order item status: %w", err)
	}
	var coup models.Coupon
	if order.CouponCode!=""{
		if err:=database.DB.Model(&coup).Where("code=?",order.CouponCode).First(&coup).Error;err!=nil{
			return 0,fmt.Errorf("failed to retrieve coupon items: %w", err)
		}
		cancelprice=ordItem.Price-(ordItem.Price-(ordItem.Price*(coup.DiscountPercentage/100)))
	}
	cancelprice=ordItem.Price
	var remainingItems []models.OrderItem
	if err := database.DB.Where("order_id = ? AND order_status != ?", orderID, models.Return).Find(&remainingItems).Error; err != nil {
		return 0,fmt.Errorf("failed to retrieve remaining order items: %w", err)
	}
	if err:=database.DB.Where("order_status !=?",models.Cancelled).Find(&remainingItems).Error;err!=nil{
		return 0,fmt.Errorf("failed to retrieve remaining order items: %w", err)
	}
	newTotal := 0.0
	for _, item := range remainingItems {
		newTotal += item.Price
	}
	
	if err := database.DB.Model(&models.Order{}).Where("order_id = ?", orderID).Update("total", newTotal).Error; err != nil {
		return 0,fmt.Errorf("failed to update order total: %w", err)
	}
	


	return cancelprice,nil
}