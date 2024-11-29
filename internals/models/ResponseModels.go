package models

import "time"

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
type UserResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"` // Change to string
	Picture      string `json:"picture"`
	ReferralCode string `json:"referral_code"`
	LoginMethod  string `json:"login_method"`
	Blocked      bool   `json:"blocked"`
}
type CartProduct struct {
	Product_id  uint	``
	Name        string  `validate:"required" json:"name"`
	Description string  `gorm:"column:description" validate:"required" json:"description"`
	CategoryID  uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Price       float64 `validate:"required,number" json:"price"`
	FinalPrice float64 `validate:"required,number" json:"final_price"`
	Quantity    uint    `validate:"required" json:"quantity"`
	ImageURL    string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}
type OrderResponse struct {
	OrderID         uint                `json:"order_id"`
	OrderDate       time.Time           `json:"order_date"`
	SubtotalAmount  float64             `json:"subtotal_amount"`  // RawAmount
	OfferTotal      float64             `json:"offer_total"` 
	DiscountAmount  float64             `json:"discount_amount"`  // DiscountPrice     // Total after discounts, renamed for consistency
	ShippingCharge  float64             `json:"shipping_charge"`  // DeliveryCharge
	TotalPayable    float64             `json:"total_payable"`    // FinalAmount
	ShippingAddress ShippingAddress     `gorm:"type:json" json:"shipping_address"`
	OrderStatus     string              `json:"order_status"`
	PaymentStatus   string              `json:"payment_status"`
	PaymentMethod   string              `json:"payment_method"`
	Items           []OrderItemResponse `json:"items"`
}

type  OrderItemResponse struct {
	OrderID   uint `json:"order_id"`
	ProductId uint  `json:"product_id"`
	ProductName  string `json:"product_name"`
	ImageURL     string `json:"image_url"`
	CategoryId    uint `json:"category_id"`
	Description    string `json:"description"`
	Price          float64 `json:"price"`
	FinalAmount   float64 `json:"final_amount"`
	Quantity       uint `json:"quantity"`
	TotalPrice     float64 `json:"total_price"`
	OrderStatus     string `json:"order_status"`
}
type AdminOrderResponse struct {
	UserId        uint `json:"user_id"`
	UserName       string `json:"user_name"`
	OrderID   uint      `json:"order_id"`
	OrderDate time.Time `json:"order_date"`
	OfferTotal float64 `json:"total" gorm:"column:total"`
	DiscountPrice  float64 `json:"discount_price"`
	FinalAmount   float64 `json:"final_amount"`
	ShippingAddress ShippingAddress  `gorm:"type:json" json:"shipping_address"`
	OrderStatus  string `json:"order_status"`
	PaymentStatus  string `json:"payment_status"`
	PaymentMethod   string `json:"payment_method"`
	Items         []OrderItemResponse `json:"items"`
}
type ResponseWhishlist struct{
	ProductID uint `json:"product_id"`
	ProductName string `json:"product_name"`
	ProductImage string `json:"product_image"`
	ProductDescription string `json:"product_description"`
	ProductPrice float64 `json:"product_price"`
	OfferPrice  float64 `json:"offer_price"`
}
type OrderCount struct {
	TotalOrder     uint `json:"total_order"`
	TotalPending   uint `json:"total_pending"`
	TotalConfirmed uint` json:"total_confirmed"`
	TotalShipped   uint `json:"total_shipped"`
	TotalDelivered uint `json:"total_delivered"`
	TotalCancelled uint `json:"total_cancelled"`
	TotalReturned  uint `json:"total_returned"`
}

type AmountInformation struct {
	TotalAmountBeforeDeduction  float64 `json:"total_amount_before_deduction"`
	TotalCouponDeduction        float64 `json:"total_coupon_deduction"`
	//TotalCategoryOfferDeduction float64 `json:"total_category_offer_deduction"`
	TotalProuctOfferDeduction   float64 `json:"total_product_offer_deduction"`
	//TotalDeliveryCharges        float64 `json:"total_delivery_charge"`
	TotalAmountAfterDeduction   float64 `json:"total_amount_after_deduction"`
}
type WalletResponse struct{
	WalletPaymentId string `json:"wallet_payment_id" gorm:"column:wallet_payment_id"`
	Type string 
	Amount uint `json:"amount" gorm:"column:amount"`
	OrderId uint `json:"order_id" gorm:"column:order_id"`
	TransactionTime time.Time `json:"transaction_time" gorm:"column:transaction_time"`
	CurrentBalance uint `json:"current_balance" gorm:"column:current_balance"`
	Reason string `json:"reason" gorm:"column:reason"`
}
type OrderDetails struct {
	CustomerName    string
	CustomerAddress ShippingAddress
	CustomerCity    string
	OrderDate time.Time
	Items           []InvoiceItem
		
}
type InvoiceItem struct{
	Name string `json:"name"`
	Quantity uint `json:"quantity"`
	Price float64 `json:"price"`
}
type BestSellingProduct struct {
    ProductID   uint   `json:"product_id"`
    ProductName string `json:"product_name"`
    TotalSold   int    `json:"total_sold"`
}

type BestSellingCategory struct {
    CategoryID   uint   `json:"category_id"`
    CategoryName string `json:"category_name"`
    TotalSold    int    `json:"total_sold"`
}
type BookingResponse struct {
	UserID uint `json:"user_id"`
	BookingId uint `json:"booking_id"`
	TimeSlot string `json:"time_slot" `
	BookingDate time.Time `json:"booking_date" `
	BookingStatus string `json:"booking_status" `
	Amount float64 `json:"amount_paid"`
	PaymentMethod string `json:"payment_method" `
	PaymentStatus string `json:"payment_status" `
	Service string `json:"service" `
}
