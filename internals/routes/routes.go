package routes

import (
	"petplate/internals/controllers"
	// "net/http"
	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine){
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})
//
router.POST("/signup",controllers.SignupUser)
router.POST("/Login",controllers.EmailLogin)

router.GET("/verifyemail",controllers.VerifyEmail)

//authentication
router.GET("/resendotp",controllers.ResendOtp)

router.GET("/google/login", controllers.HandleGoogleLogin)
router.GET("/auth/callback", controllers.HandleGoogleCallback)
router.POST("admin/login",controllers.AdminLogin)


}
