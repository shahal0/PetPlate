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
	Quantity    uint    `validate:"required" json:"quantity"`
	ImageURL    string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}
type OrderResponse struct {
	OrderID   uint      `json:"order_id"`
	OrderDate time.Time `json:"order_date"`
	RawAmount  float64 `json:"raw_amount"`
	OfferTotal float64 `json:"total" gorm:"column:total"`
	DiscountPrice  float64 `json:"discount_price"`
	FinalAmount   float64 `json:"final_amount"`
	ShippingAddress ShippingAddress  `gorm:"type:json" json:"shipping_address"`
	OrderStatus  string `json:"order_status"`
	PaymentStatus  string `json:"payment_status"`
	PaymentMethod   string `json:"payment_method"`
	Items         []OrderItemResponse `json:"items"`
}
type  OrderItemResponse struct {
	OrderID   uint `json:"order_id"`
	ProductId uint  `json:"product_id"`
	ProductName  string `json:"product_name"`
	ImageURL     string `json:"image_url"`
	CategoryId    uint `json:"category_id"`
	Description    string `json:"description"`
	Price          float64 `json:"price"`
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

