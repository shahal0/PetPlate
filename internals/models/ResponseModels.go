package models


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
    PhoneNumber  string `json:"phone_number"`  // Change to string
    Picture      string `json:"picture"`
    ReferralCode string `json:"referral_code"`
    LoginMethod  string `json:"login_method"`
    Blocked      bool   `json:"blocked"`
}
