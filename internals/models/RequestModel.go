package models

import (

	"github.com/dgrijalva/jwt-go"
)

type SignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     uint   `validate:"required,number,min=1000000000,max=9999999999" json:"phone_number"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}
type EmailLoginRequest struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required" json:"password"`
}
type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string
}
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
type AddProductRequest struct {
	ID          uint
	Name        string  `validate:"required" json:"name"`
	Description string  `gorm:"column:description" validate:"required" json:"description"`
	CategoryID  uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Price       float64 `validate:"required,number" json:"price"`
	OfferPrice        float64 `validate:"required,number"gorm:"column:offer_price" json:"offer_price"`
	MaxStock    uint    `validate:"required,number" json:"max_stock"`
	RatingSum   float64 `gorm:"column:rating_sum" json:"rating_sum"`
	ImageURL    string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}
type ServiceRequest struct {
	ID          uint
	Name        string  `validate:"required" json:"name"`
	Description string  `gorm:"column:description" validate:"required" json:"description"`
	Price       float64 `validate:"required,number" json:"price"`
	ImageURL    string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}
type UserRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Picture     string `json:"picture"`
}
type AddressRequest struct {
	UserID       uint   `json:"user_id" gorm:"column:user_id"`
	AddressID    uint   `gorm:"primaryKey;autoIncrement;column:address_id" json:"address_id"`
	PhoneNumber  uint   `gorm:"column:phone_number" validate:"required,number,min=1000000000,max=9999999999" json:"phone_number"`
	AddressType  string `validate:"required" json:"address_type" gorm:"column:address_type"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}
type CartRequest struct {
	ProductID uint `json:"product_id" gorm:"column:product_id"`
	Quantity  uint `validate:"required,number" json:"quantity"`
	CategoryOffer float64  `json:"category_offer"`
}
type OrderRequest struct {
	PaymentMethod uint `json:"payment_method" validate:"required"`
	AddressID     uint `json:"address_id" validate:"required"`
}
type RatingRequest struct {
	ProductID uint    `json:"product_id" validate:"required"`
	Rating    float64 `json:"rating" validate:"required,min=1,max=5"`
	Comment   string  `json:"comment"`
}
type PasswordChange struct {
	OldPassword     string `json:"old_password" validate:"required,min=8,max=128"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,max=128`
}
type CouponRequest struct{
	Code  string `json:"code" validate:"required,min=3,max=10"`
	DiscountPercentage float64   `json:"discount_percentage"validate:"required,min=0,max=100"` 
	ExpirationDate    uint	`json:"expiry_date"`
	IsActive          bool      `json:"is_active"` 
	MinimumPurchase   float64   `json:"minimum_purchase"` 
	MaximumDiscountAmount  float64 `json:"maximum_discount_amount"`
}
type WhislistRequest struct {
	ProductID uint `json:"product_id" gorm:"column:product_id"`
}