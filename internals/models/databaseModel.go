package models

import (
	//"time"
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
	ID uint
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