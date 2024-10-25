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
adminRoutes := router.Group("/api/v1/admin")
adminRoutes.Use(middlewares.AuthMiddleware())
{
	adminRoutes.GET("/listusers", controllers.GetUserList) 
	log.Println("Admin routes registered")
	adminRoutes.PATCH("/blockuser", controllers.BlockUser) 
	adminRoutes.PATCH("/unblockuser", controllers.UnblockUser)
}
categoryGroup := router.Group("/admin/categories")
categoryGroup.Use(middlewares.AuthMiddleware())
    {
        categoryGroup.POST("/add", controllers.CreateCategory) // Create a new category
	   categoryGroup.PUT("/edit",controllers.CategoryEdit)
	   categoryGroup.DELETE("/delete",controllers.CategoryDelete)
    }
ProductGroup := router.Group("/admin/products")
ProductGroup.Use(middlewares.AuthMiddleware())
{
	ProductGroup.POST("/add",controllers.AddProducts)
	ProductGroup.PUT("/edit",controllers.EditProduct)
	ProductGroup.DELETE("/delete",controllers.DeleteProducts)

}
ServiceGroup := router.Group("/admin/services")
ServiceGroup.Use(middlewares.AuthMiddleware())
{
	ServiceGroup.POST("/add",controllers.AddService)
	ServiceGroup.PUT("/edit",controllers.EditService)
	ServiceGroup.DELETE("/delete",controllers.DeleteService)
}


}

