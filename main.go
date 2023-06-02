package main

import (
	"fmt"
	"net/http"

	"github.com/62teknologi/62stingray/62golib/utils"
	"github.com/62teknologi/62stingray/app/http/controllers"
	"github.com/62teknologi/62stingray/app/http/middlewares"
	"github.com/62teknologi/62stingray/config"

	"github.com/gin-gonic/gin"
)

func main() {

	configs, err := config.LoadConfig(".", &config.Data)
	if err != nil {
		fmt.Printf("cannot load config: %w", err)
		return
	}

	// todo : replace last variable with spread notation "..."
	utils.ConnectDatabase(configs.DBDriver, configs.DBSource1, configs.DBSource2)

	utils.InitPluralize()

	r := gin.Default()

	apiV1 := r.Group("/api/v1").Use(middlewares.DbSelectorMiddleware())
	{
		c := &controllers.CartController{}

		apiV1.GET("/:table", c.FindAll)
		apiV1.POST("/:table", c.Add)
		apiV1.PUT("/:table", c.Synch)
		apiV1.DELETE("/:table/:userId/:id", c.Delete)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, utils.ResponseData("success", "Server running well", nil))
	})

	err = r.Run(configs.HTTPServerAddress)

	if err != nil {
		fmt.Printf("cannot run server: %w", err)
		return
	}
}
