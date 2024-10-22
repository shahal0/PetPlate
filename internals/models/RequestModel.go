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