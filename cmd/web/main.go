package main

import (
	"backend/cmd/web/dto"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	r := gin.Default()
	err := godotenv.Load()
	if err != nil {
		log.Println(".env NOT FOUND")
	}

	httpClient := http.Client{}

	// Get the ICD API IDs etc
	icdClientID := os.Getenv("ICD_CLIENTID")
	icdClientSecret := os.Getenv("ICD_CLIENTSECRET")

	genaiClient, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Set up the dependencies
	icdRepository := repository.NewICDRepository(&httpClient, icdClientID, icdClientSecret)
	namasteRepository := repository.NewNamasteRepository()
	autocompleteService := service.NewAutoComplete(genaiClient, icdRepository, namasteRepository)

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, dto.Message{Message: "ok"})
	})

	r.GET("/sync", func(ctx *gin.Context) {
		if err := autocompleteService.Update(); err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Error{Error: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, dto.Message{Message: "All repositories synced"})
	})

	r.GET("/autocomplete", func(ctx *gin.Context) {
		query := ctx.Query("query")

		resp, err := autocompleteService.Find(ctx.Request.Context(), query)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Error{Error: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	})

	r.Run(":8000")
}
