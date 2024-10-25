package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	//"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"strings"

	"petplate/utils"
	//"os"
	"petplate/internals/database"
	"petplate/internals/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	//passwordvalidator "github.com/wagslane/go-password-validator"
	"gorm.io/gorm"
	//"gorm.io/gorm/utils"
)
var (
    googleOauthConfig = &oauth2.Config{
        RedirectURL:  "http://localhost:8080/auth/callback", // Replace with your redirect URL
        ClientID:     "1081304852347-0qtbfi9giv5f9tg6ckjrb8rul67mh8pl.apps.googleusercontent.com",               // Replace with your client ID
        ClientSecret: "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",           // Replace with your client secret
        Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
        Endpoint:     google.Endpoint,
    }
    oauthStateString = "oauthstring" // This should be randomly generated for security
)

func SignupUser(c *gin.Context) {
    var signupRequest models.SignupRequest
    if err := c.BindJSON(&signupRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "failed to process the incoming request" + err.Error(),
        })
        return
    }

    validate := validator.New()
    if err := validate.Struct(&signupRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": err.Error(),
        })
        return
    }

    if signupRequest.Password != signupRequest.ConfirmPassword {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  false,
            "message": "passwords doesn't match",
        })
        return
    }

    // Password validation (you can add your password validation logic here)

    User := models.User{
        Name:        signupRequest.Name,
        Email:       signupRequest.Email,
        PhoneNumber: fmt.Sprint(signupRequest.PhoneNumber),
        Password:    signupRequest.Password,
        LoginMethod: models.EmailLoginMethod,
        Blocked:     false,
    }

    tx := database.DB.Where("email =? AND deleted_at IS NULL", signupRequest.Email).First(&User)
    if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  false,
            "message": "failed to retrieve information from the database",
        })
        return
    } else if tx.Error == gorm.ErrRecordNotFound {
        // User does not exist, proceed to create
        tx = database.DB.Create(&User)
        if tx.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  false,
                "message": "failed to create a new user",
            })
            return
        }
    } else {
        // User already exists
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  false,
            "message": "user already exists",
        })
        return
    }

    // Generate JWT Token
    tokenString, err := utils.GenerateJWT(User.Email)
	if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "status":  false,
        "message": "failed to generate JWT token: " + err.Error(),
    })
    return
	}


    // Set the JWT token in the cookie
    c.SetCookie("Authorization", tokenString, 3600*24, "/", "localhost", false, true)

    // Send OTP (as per your original implementation)
    VerificationTable := models.VerificationTable{
        Email:              User.Email,
        Role:               models.UserRole,
        VerificationStatus: models.VerificationStatusPending,
    }
    SendOtp(c, User.Email, VerificationTable.OTPExpiry)

    // Return success response
    c.JSON(http.StatusOK, gin.H{
        "status":  true,
        "message": "Email login successful, please login to complete your email verification",
        "data": gin.H{
            "user": gin.H{
                "name":         User.Name,
                "email":        User.Email,
                "phone_number": User.PhoneNumber,
                "picture":      User.Picture,
                "login_method": User.LoginMethod,
                "block_status": User.Blocked,
            },
        },
    })
}

func GenerateJWT(s string) {
	panic("unimplemented")
}
func EmailLogin(c *gin.Context) {
	//get the json from the request
	var LoginRequest models.EmailLoginRequest
	if err := c.BindJSON(&LoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process the incoming request",
		})
		return
	}
	validate := validator.New() // Initialize the validator
	if err := validate.Struct(&LoginRequest); err != nil {
		// Validation failed, return error message
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	
	var user models.User
	tx := database.DB.Where("email =? AND deleted_at IS NULL", LoginRequest.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
		})
		return
	}
	if user.LoginMethod != models.EmailLoginMethod {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "email uses another method for logging in, use google sso",
		})
		return
	}
	tokenString, err := utils.GenerateJWT(user.Email)
	if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "status":  false,
        "message": "failed to generate JWT token: " + err.Error(),
    })
    return
	}


    // Set the JWT token in the cookie
    c.SetCookie("Authorization", tokenString, 3600*24, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email login successful.",
		"data": gin.H{
			"user": gin.H{
				"name":         user.Name,
				"email":        user.Email,
				"phone_number": user.PhoneNumber,
				"picture":      user.Picture,
				"login_method": user.LoginMethod,
				"block_status": user.Blocked,
				"token":tokenString,
			},
		},
	})
}
func SendOtp(c *gin.Context, to string, otpexpiry uint64) error {
	// Random OTP generation
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000 // Generates a 6-digit OTP
	now := time.Now().Unix()       // Get current time in Unix format

	// If the OTP expiry is still valid, return a message
	if otpexpiry > 0 && uint64(now) < otpexpiry {
		timeLeft := otpexpiry - uint64(now)
		return errors.New(fmt.Sprintf("OTP is still valid. Please wait %v seconds before sending another request.", int(timeLeft)))
	}
	fmt.Println("hi")

	// Set OTP expiry time (5 minutes from now)
	expiryTime := now + 5*60 // 5 minutes in seconds

	// Email configuration
	from := "petplate0@gmail.com"
	appPassword := "yjefvtzmglurwcid" //os.Getenv("SMTPAPP") // Fetch SMTP app password from environment variable
	fmt.Println(appPassword)
	if appPassword == "" {
		fmt.Println("es")
		return errors.New("SMTP app password not set")
	}

	// SMTP authentication and message setup
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")
	otpStr := strconv.Itoa(otp) // Convert OTP to string
	message := []byte("Subject: Your OTP Code\n\nYour OTP is: " + otpStr)
	fmt.Println("23")

	// Send email
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, message)
	if err != nil {
		fmt.Println("Failed to send email:", err) // Print the error message for debugging
		return errors.New("failed to send email: " + err.Error())
	}

	// Create the VerificationTable record for storing the OTP
	VerificationTable := models.VerificationTable{
		Email:              to,
		OTP:                uint64(otp),
		OTPExpiry:          uint64(expiryTime),
		VerificationStatus: models.VerificationStatusPending,
	}

	// Use FirstOrCreate to insert if the record doesn't exist, otherwise update it
	if err := database.DB.Where("email = ?", VerificationTable.Email).
	Assign(VerificationTable).
	FirstOrCreate(&VerificationTable).Error; err != nil {
	return errors.New("failed to store OTP information in the database: " + err.Error())
	}
	return nil
}
func VerifyEmail(c *gin.Context) {
	// Extract query parameters from the URL
	entityEmail := c.Query("email")
	entityOTP := c.Query("otp")

	// Trim any leading/trailing spaces
	entityEmail = strings.TrimSpace(entityEmail)
	entityOTP = strings.TrimSpace(entityOTP)

	// Validate that email and OTP are provided
	if entityEmail == "" || entityOTP == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Failed to process incoming request",
		})
		return
	}

	// Convert OTP to an integer
	OTP, err := strconv.Atoi(entityOTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid OTP format",
		})
		return
	}

	// Query the verification table for the given email
	var VerificationTable models.VerificationTable
	tx := database.DB.Where("email = ?", entityEmail).First(&VerificationTable)

	// Check if the record is not found
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Email not found",
		})
		return
	}

	// Handle other errors
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to retrieve information",
		})
		return
	}

	// Check if OTP is already used
	if VerificationTable.OTP == 0 {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  false,
			"message": "Please login again to verify your email",
		})
		return
	}

	// Check if OTP has expired
	if VerificationTable.OTPExpiry < uint64(time.Now().Unix()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "OTP has expired, please login again to verify your OTP",
		})
		return
	}

	// Check if the provided OTP is correct
	if VerificationTable.OTP != uint64(OTP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid OTP",
		})
		return
	}

	// Update the verification status to "verified"
	VerificationTable.VerificationStatus = models.VerificationStatusVerified
	tx = database.DB.Where("email = ?", entityEmail).Updates(&VerificationTable)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to verify email, please try again",
		})
		return
	}

	// Generate a JWT token for the user
	tokenString, err := utils.GenerateJWT(entityEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to generate JWT token",
		})
		return
	}

	// Return the success message and JWT token
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email verification successful",
		"data": gin.H{
			"token": tokenString, // Send the JWT token as part of the response
		},
	})
}

func ResendOtp(c *gin.Context) {
	// Get the email from the request
	entityEmail := c.Query("email")
	if entityEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Email is required",
		})
		return
	}

	var verification models.VerificationTable
	// Fetch the existing OTP details for the email
	tx := database.DB.Where("email = ?", entityEmail).First(&verification)

	// Check if the record exists
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Email not found",
		})
		return
	}

	// Check if the OTP is still valid
	currentTime := uint64(time.Now().Unix())
	if verification.OTPExpiry > currentTime {
		// If OTP is still valid, prevent resend and inform the user
		timeLeft := verification.OTPExpiry - currentTime
		c.JSON(http.StatusTooManyRequests, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("OTP is still valid. Please wait %v seconds before requesting another OTP.", timeLeft),
		})
		return
	}

	// OTP has expired, generate and send a new OTP using the existing SendOtp function
	if err := SendOtp(c, entityEmail, verification.OTPExpiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to send OTP: " + err.Error(),
		})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "A new OTP has been sent to your email.",
	})
}
func HandleGoogleLogin(c *gin.Context) {
	//utils.NoCache(c)
	url := googleOauthConfig.AuthCodeURL("hjdfyuhadVFYU6781235")
	c.Redirect(http.StatusTemporaryRedirect, url)
	c.Next()
}
func HandleGoogleCallback(c *gin.Context) {
    // Get the authorization code from Google
    fmt.Println("Starting to handle callback")
    code := c.Query("code")
    fmt.Println("Authorization Code:", code)

    // Check if code exists in the query parameters
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "missing code parameter",
        })
        return
    }

    // Exchange authorization code for access token
    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failed",
            "message": "failed to exchange token",
        })
        return
    }

    // Fetch user information from Google using the access token
    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status": "failed",
            "message": "failed to get user information",
        })
        return
    }
    defer response.Body.Close()

    // Read the response body
    content, err := io.ReadAll(response.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "failed to read user information",
        })
        return
    }

    // Unmarshal the Google response into a struct
    var googleUser models.GoogleResponse
    err = json.Unmarshal(content, &googleUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failed",
            "message": "failed to parse user information",
        })
        return
    }

    // Prepare the new user struct with Google data
    newUser := models.User{
        Name:        googleUser.Name,
        Email:       googleUser.Email,
        LoginMethod: models.GoogleSSOMethod, // Assuming GoogleSSOMethod is defined in your models
        Picture:     googleUser.Picture,
        Blocked:     false,
    }

    // If the user's name is missing, fallback to their email as the name
    if newUser.Name == "" {
        newUser.Name = googleUser.Email
    }

    // Check if the user already exists in the database
    var existingUser models.User
    if err := database.DB.Where("email = ? AND deleted_at IS NULL", newUser.Email).First(&existingUser).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // User not found, create a new user
            if err := database.DB.Create(&newUser).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "status":  "failed",
                    "message": "failed to create user through Google SSO",
                })
                return
            }
        } else {
            // Failed to fetch user due to another error
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "failed",
                "message": "failed to fetch user information",
            })
            return
        }
    }
	tokenString, err := utils.GenerateJWT(newUser.Email)
	if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "status":  false,
        "message": "failed to generate JWT token: " + err.Error(),
    })
    return
	}


    // Set the JWT token in the cookie
    c.SetCookie("Authorization", tokenString, 3600*24, "/", "localhost", false, true)

    // If no error occurred, respond with success
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "user authenticated successfully",
        "user":    newUser,
		"token":tokenString,
    })
}



