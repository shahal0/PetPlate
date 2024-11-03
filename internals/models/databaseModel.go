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
}
type Product struct{
	gorm.Model
	ID uint			`gorm:"primaryKey;autoIncrement;column:product_id" json:"product_id"`
	Name  			string `validate:"required" json:"name"`
	Description     string  `gorm:"column:description" validate:"required" json:"description"`
	CategoryID      uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Price           float64 `validate:"required,number" json:"price"`
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
	Total float64 `json:"total" gorm:"column:total"`
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




	
	
	
