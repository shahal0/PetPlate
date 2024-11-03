package routes

import (
	"log"
	"petplate/internals/controllers"
	"petplate/internals/middlewares"

	//"petplate/utils"

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
router.POST("/user/signup",controllers.SignupUser)
router.POST("/user/login",controllers.EmailLogin)

router.GET("/verifyemail",controllers.VerifyEmail)

//authentication
router.GET("/resendotp",controllers.ResendOtp)

router.GET("/google/login", controllers.HandleGoogleLogin)
router.GET("/auth/callback", controllers.HandleGoogleCallback)
router.POST("admin/login",controllers.AdminLogin)
router.GET("/listproduct",controllers.ListProduct)
router.GET("/listservice",controllers.GetServices)
router.GET("/listcategories", controllers.GetCategories)
router.GET("/listproductbycategory", controllers.CategoryByProduct)
router.GET("/user/profie",controllers.GetUserProfile)
adminRoutes := router.Group("/admin")
adminRoutes.Use(middlewares.AuthMiddleware("admin"))
{
	adminRoutes.GET("/listusers", controllers.GetUserList) 
	log.Println("Admin routes registered")
	adminRoutes.PATCH("/blockuser", controllers.BlockUser) 
	adminRoutes.PATCH("/unblockuser", controllers.UnblockUser)
	adminRoutes.GET("/orders",controllers.AdminOrderList)
	adminRoutes.PATCH("/orders/cancel",controllers.AdminCancelOrder)
	adminRoutes.PATCH("/orders/cancel/item",controllers.CancelItemFromAdminOrders)
	adminRoutes.PATCH("/orders/updatestatus",controllers.UpdateOrderstatus)

}
categoryGroup := router.Group("/admin/categories")
categoryGroup.Use(middlewares.AuthMiddleware("admin"))
    {
        categoryGroup.POST("/add", controllers.CreateCategory) // Create a new category
	   categoryGroup.PUT("/edit",controllers.CategoryEdit)
	   categoryGroup.DELETE("/delete",controllers.CategoryDelete)
    }
ProductGroup := router.Group("/admin/products")
ProductGroup.Use(middlewares.AuthMiddleware("admin"))
{
	ProductGroup.POST("/add",controllers.AddProducts)
	ProductGroup.PUT("/edit",controllers.EditProduct)
	ProductGroup.DELETE("/delete",controllers.DeleteProducts)

}
ServiceGroup := router.Group("/admin/services")
ServiceGroup.Use(middlewares.AuthMiddleware("admin"))
{
	ServiceGroup.POST("/add",controllers.AddService)
	ServiceGroup.PUT("/edit",controllers.EditService)
	ServiceGroup.DELETE("/delete",controllers.DeleteService)
}
UserRoutes := router.Group("/user")
UserRoutes.Use(middlewares.AuthMiddleware("user"))
{
	UserRoutes.GET("/profile",controllers.GetUserProfile)
	UserRoutes.PUT("/profile/edit",controllers.EditProfile)
	UserRoutes.PUT("/profile/changepassword",controllers.ChangePassword)
}
AddressGroup := router.Group("/user/address")
AddressGroup.Use(middlewares.AuthMiddleware("user"))
{
	AddressGroup.POST("/add",controllers.AddAddress)
	AddressGroup.PUT("/edit",controllers.EditAddress)
	AddressGroup.DELETE("/delete",controllers.DeleteAddress)
}
CartGroup := router.Group("/user/cart")
CartGroup.Use(middlewares.AuthMiddleware("user"))
{
	CartGroup.POST("/add",controllers.AddCart)
	CartGroup.GET("/list",controllers.ListCart)
	CartGroup.DELETE("/delete",controllers.DeleteFromCart)
}
OrderGroup := router.Group("/user/order")
OrderGroup.Use(middlewares.AuthMiddleware("user"))
{
	OrderGroup.POST("/placeorder",controllers.PlaceOrder)
	OrderGroup.GET("/seeorders",controllers.UserSeeOrders)
	OrderGroup.PATCH("/cancel",controllers.UserCancelOrder)
	OrderGroup.PATCH("/cancel/item",controllers.CancelItemFromUserOrders)
	OrderGroup.PUT("/product/rating",controllers.ProductRating)
	OrderGroup.GET("/product/search",controllers.SearchProduct)
}

}

