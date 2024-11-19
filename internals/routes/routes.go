package routes

import (
	"log"
	//"petplate/internals/controllers"
	"petplate/internals/controllers"
	"petplate/internals/middlewares"

	//"petplate/utils"

	// "net/http"
	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})

	//

	router.POST("/user/signup", controllers.SignupUser)
	router.POST("/user/login", controllers.EmailLogin)

	router.GET("/verifyemail", controllers.VerifyEmail)

	//authentication
	router.GET("/resendotp", controllers.ResendOtp)

	router.GET("/google/login", controllers.HandleGoogleLogin)
	router.GET("/auth/callback", controllers.HandleGoogleCallback)
	router.POST("admin/login", controllers.AdminLogin)
	router.GET("/listproduct", controllers.ListProduct)
	router.GET("/listservice", controllers.GetServices)
	router.GET("/listcategories", controllers.GetCategories)
	router.GET("/listproductbycategory", controllers.CategoryByProduct)
	router.GET("/user/profie", controllers.GetUserProfile)
	router.GET("/payment", controllers.RenderRayzorPay)
	router.POST("/create-order", controllers.CreateOrder)
	router.POST("/verify-payment/:orderid", controllers.VerifyPayment)
	router.POST("/payment-failed", controllers.FailedHandling)
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middlewares.AuthMiddleware("admin"))
	{
		adminRoutes.GET("/listusers", controllers.GetUserList)
		log.Println("Admin routes registered")
		adminRoutes.PATCH("/blockuser", controllers.BlockUser)
		adminRoutes.PATCH("/unblockuser", controllers.UnblockUser)
		adminRoutes.GET("/orders", controllers.AdminOrderList)
		adminRoutes.PATCH("/orders/cancel", controllers.AdminCancelOrder)
		adminRoutes.PATCH("/orders/cancel/item", controllers.CancelItemFromAdminOrders)
		adminRoutes.PATCH("/orders/updatestatus", controllers.UpdateOrderstatus)

	}
	categoryGroup := router.Group("/admin/categories")
	categoryGroup.Use(middlewares.AuthMiddleware("admin"))
	{
		categoryGroup.POST("/add", controllers.CreateCategory)
		categoryGroup.PUT("/edit", controllers.CategoryEdit)
		categoryGroup.DELETE("/delete", controllers.CategoryDelete)
	}
	salesGroup:=router.Group("/admin/salesReport")
	salesGroup.Use(middlewares.AuthMiddleware("admin"))
	{
		salesGroup.GET("/get",controllers.SalesReport)
		salesGroup.GET("/download",controllers.DownloadSalesReportPDF)
		salesGroup.GET("/BestSellingproduct",controllers.BestSellingProduct)
		salesGroup.GET("/BestSellingcategory",controllers.BestsellingCategory)
	}
	ProductGroup := router.Group("/admin/products")
	ProductGroup.Use(middlewares.AuthMiddleware("admin"))
	{
		ProductGroup.POST("/add", controllers.AddProducts)
		ProductGroup.PUT("/edit", controllers.EditProduct)
		ProductGroup.DELETE("/delete", controllers.DeleteProducts)

	}
	CouponGroup := router.Group("/admin/coupon")
	CouponGroup.Use(middlewares.AuthMiddleware("admin"))
	{
		CouponGroup.POST("/add", controllers.AdminAddCoupon)
		CouponGroup.PATCH("/disable", controllers.DisableCoupon)
		CouponGroup.GET("/show", controllers.ListCoupon)
	}
	ServiceGroup := router.Group("/admin/services")
	ServiceGroup.Use(middlewares.AuthMiddleware("admin"))
	{
		ServiceGroup.POST("/add", controllers.AddService)
		ServiceGroup.PUT("/edit", controllers.EditService)
		ServiceGroup.DELETE("/delete", controllers.DeleteService)
	}
	UserRoutes := router.Group("/user")
	UserRoutes.Use(middlewares.AuthMiddleware("user"))
	{
		UserRoutes.GET("/profile", controllers.GetUserProfile)
		UserRoutes.PUT("/profile/edit", controllers.EditProfile)
		UserRoutes.PUT("/profile/changepassword", controllers.ChangePassword)
		UserRoutes.GET("/wallet",controllers.WalletHistory)
	}
	AddressGroup := router.Group("/user/address")
	AddressGroup.Use(middlewares.AuthMiddleware("user"))
	{
		AddressGroup.POST("/add", controllers.AddAddress)
		AddressGroup.PUT("/edit", controllers.EditAddress)
		AddressGroup.DELETE("/delete", controllers.DeleteAddress)
	}
	CartGroup := router.Group("/user/cart")
	CartGroup.Use(middlewares.AuthMiddleware("user"))
	{
		CartGroup.POST("/add", controllers.AddCart)
		CartGroup.GET("/list", controllers.ListCart)
		CartGroup.DELETE("/delete", controllers.DeleteFromCart)
	}
	OrderGroup := router.Group("/user/order")
	OrderGroup.Use(middlewares.AuthMiddleware("user"))
	{
		OrderGroup.POST("/placeorder", controllers.PlaceOrder)
		OrderGroup.GET("/seeorders", controllers.UserSeeOrders)
		OrderGroup.PATCH("/cancel", controllers.UserCancelOrder)
		OrderGroup.PATCH("/cancel/item", controllers.CancelItemFromUserOrders)
		OrderGroup.PUT("/product/rating", controllers.ProductRating)
		OrderGroup.GET("/product/search", controllers.SearchProduct)
		OrderGroup.PATCH("/product/return",controllers.ReturnOrder)
		OrderGroup.GET("/invoice",controllers.Invoice)
	}
	WhislistGroup := router.Group("/user/whishlist")
	WhislistGroup.Use(middlewares.AuthMiddleware("user"))
	{
		WhislistGroup.POST("/add", controllers.AddToWhishlist)
		WhislistGroup.GET("/list", controllers.SowWhislist)
		WhislistGroup.PATCH("/addtocart", controllers.AddToCart)
	}

}
