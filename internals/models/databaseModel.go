package models

import (
	//"time"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uint    `validate:"required"`
	Name           string  `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string  `gorm:"column:email;type:varchar(255);unique_index" validate:"email" json:"email"`
	PhoneNumber    string  `gorm:"column:phone_number;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture        string  `gorm:"column:picture;type:text" json:"picture"`
	ReferralCode   string  `gorm:"column:referral_code" json:"referral_code"`
	// WalletAmount   float64 `gorm:"column:wallet_amount;type:double" json:"wallet_amount"`
	LoginMethod    string  `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	Blocked        bool    `gorm:"column:blocked;type:bool" json:"blocked"`
	Password       string	`gorm:"column:password;type:varchar(255)" validate:"required,min=8" json:"password"`
	// Salt           string  `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	HashedPassword string  `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
}
type VerificationTable struct {
	Email              string `validate:"required,email" gorm:"type:varchar(255);unique_index"`
	Role               string
	OTP                uint64
	OTPExpiry          uint64
	VerificationStatus string `gorm:"type:varchar(255)"`
}
type Admin struct {
	gorm.Model
	Email string `validate:"required,email"`
	Password string
}
type Category struct {
    gorm.Model
    Name        string `json:"name" gorm:"unique;not null"`
    Description string `json:"description"`
	CategoryOffer float64  `json:"category_offer"`
}
type Product struct{
	ID uint			`gorm:"primaryKey;autoIncrement;column:product_id" json:"product_id"`
	Name  			string `validate:"required" json:"name"`
	Description     string  `gorm:"column:description" validate:"required" json:"description"`
	CategoryID      uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Price           float64 `validate:"required,number" json:"price"`
	OfferPrice        float64 `gorm:"column:offer_price" json:"offer_price"`
	MaxStock        uint    `validate:"required,number" json:"max_stock"`
	RatingSum       float64 `gorm:"column:rating_sum" json:"rating_sum"`
	ImageURL        string  `gorm:"column:image_url" validate:"required" json:"image_url"`

}
type Service struct{
	ID 		uint
	Name 	string `validate:"required" json:"name"`
	Description string	`gorm:"column:description" validate:"required" json:"description"`
	Price   float64	`validate:"required,number" json:"price"`
	ImageURL string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}
type Address struct {
	UserID       uint   `json:"user_id" gorm:"column:user_id"`
	AddressID    uint   `gorm:"primaryKey;autoIncrement;column:address_id" json:"address_id"`
	PhoneNumber  uint   `gorm:"column:phone_number" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
	AddressType  string `validate:"required" json:"address_type" gorm:"column:address_type"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}
type Cart struct{
	UserID uint   `json:"user_id" gorm:"column:user_id"`
	ProductID uint	`json:"product_id"`
	Quantity uint
}
type ShippingAddress struct{
	PhoneNumber  uint   `gorm:"column:phone_number" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
	AddressType  string `validate:"required" json:"address_type" gorm:"column:address_type"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}
type Order  struct{
	OrderID uint `gorm:"primaryKey;autoIncrement;column:order_id" json:"order"`
	UserID uint `json:"user_id" gorm:"column:user_id"`
	OrderDate time.Time `json:"order_date" gorm:"column:order_date"`
	RawAmount  float64 `json:"raw_amount" gorm:"column:raw_amount"`
	OfferTotal float64 `json:"total" gorm:"column:total"`
	CouponCode	string  `json:"coupon_code" gorm:"column:coupon_code"`
	DiscountAmount  float64 `json:"discount_amount" gorm:"column:discount_amount"`
	FinalAmount  float64 `json:"final_amount" gorm:"column:final_amount"`
	PaymentMethod  string  `json:"payment_method" gorm:"column:payment_method"`
	ShippingAddress ShippingAddress `gorm:"embedded" json:"shipping_address"`
	PaymentStatus   string `json:"payment_status" gorm:"column:payment_status"`
	OrderStatus string `json:"order_status" gorm:"column:order_status"`

}
type  OrderItem struct{
	OrderItemID  uint `gorm:"primaryKey;autoIncrement" json:"order_item_id"`
	OrderID uint `json:"order_id" gorm:"column:order_id"`
	ProductID uint `json:"product_id" gorm:"column:product_id"`
	Quantity uint `json:"quantity" gorm:"column:quantity"`
	Price  float64 `json:"price" gorm:"column:price"`
	OrderStatus  string `json:"order_status" gorm:"column:order_status"`
}
type Rating  struct{
	RatingID uint `gorm:"primaryKey;autoIncrement" json:"rating_id"`
	ProductID  uint `json:"product_id" gorm:"column:product_id"`
	Rating   float64 `json:"rating" gorm:"column:rating"`
	Comment   string `json:"comment" gorm:"column:comment"`

}
type Payment struct{
	PaymentID uint `gorm:"primaryKey;autoIncrement" json:"payment_id" `
	OrderID string `json:"order_id" gorm:"column:order_id"`
	RayzorpayOrderID  string `json:"rayzorpay_order_id" gorm:"column:rayzorpay`
	RayzorPayPaymentID  string `json:"rayzorpay_payment_id" gorm:"column:rayzorpay`
	RayzorPaySignature   string `json:"rayzorpay_signature" gorm:"column:rayzorpay`
	PaymentGateway  string `json:"payment_gateway" gorm:"column:payment_gateway"`
	PaymentStatus    string `json:"payment_status" gorm:"column:payment_status"`
	AmountPaid 	float64`json:"amount_paid" gorm:column:amount_paid`
}
type Coupon struct {
	Code              string    `gorm:"type:varchar(255);unique_index" json:"code"`
	DiscountPercentage float64   `json:"discount_percentage"` 
	ExpirationDate    time.Time `json:"expiration_date"`
	IsActive          bool      `json:"is_active"` 
	MinimumPurchase   float64   `json:"minimum_purchase"` 
	Description       string    `json:"description"` 
	MaximumDiscountAmount  float64 `json:"maximum_discount_amount"`
}
type CouponUsage struct{
	Coupon string  `json:"coupon" gorm:"column:coupon"`
	UserID  uint `json:"user_id" gorm:"column:user_id"`
	UsageCOunt  uint `json:"usage_count" gorm:"column:usage_count"`

}
type Whishlist struct{
	UserID uint   `json:"user_id" gorm:"column:user_id"`
	ProductID uint	`json:"product_id"`
}






	
	
	
