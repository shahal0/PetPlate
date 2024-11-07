package main

import (
	"petplate/internals/database"
	"petplate/internals/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectToDB()

	router := gin.Default()

	routes.InitRoutes(router)
	router.LoadHTMLGlob("templates/*.html")
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
