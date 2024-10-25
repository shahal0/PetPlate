package models
import "github.com/dgrijalva/jwt-go"

type SignupRequest struct{
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     uint   `validate:"required,number,min=1000000000,max=9999999999" json:"phone_number"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}
type EmailLoginRequest struct{
	Email           string `validate:"required,email" json:"email"`
	Password        string `validate:"required" json:"password"`
}
type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}
type AdminLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Password string 
}
type CreateCategoryRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}
type AddProductRequest struct{
	ID uint
	Name  			string `validate:"required" json:"name"`
	Description     string  `gorm:"column:description" validate:"required" json:"description"`
	CategoryID      uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Price           float64 `validate:"required,number" json:"price"`
	MaxStock        uint    `validate:"required,number" json:"max_stock"`
	RatingSum       float64 `gorm:"column:rating_sum" json:"rating_sum"`
	ImageURL        string  `gorm:"column:image_url" validate:"required" json:"image_url"`

}
type ServiceRequest struct{
	ID uint
	Name 	string `validate:"required" json:"name"`
	Description string	`gorm:"column:description" validate:"required" json:"description"`
	Price   float64	`validate:"required,number" json:"price"`
	ImageURL string  `gorm:"column:image_url" validate:"required" json:"image_url"`
}