package main

import (
	"backend/cmd/web/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	r := gin.Default()
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, dto.Message{Message: "ok"})
	})

	r.Run(":8000")
}
