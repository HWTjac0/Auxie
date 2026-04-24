package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{

		}
	}
}
