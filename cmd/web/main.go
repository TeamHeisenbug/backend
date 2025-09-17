package main

import (
	"backend/cmd/web/dto/controller"
	"backend/docs"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"google.golang.org/genai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env NOT FOUND")
	}

	host := os.Getenv("RENDER_EXTERNAL_HOSTNAME")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r := gin.Default()
	r.Use(cors.Default())

	docs.SwaggerInfo.Title = "NEXUS API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "NEXUS (NAMASTE - ICD Exchange for Unified Standards) is a smart, FHIR R4 - compliant service. It connect India's NAMASTE codes for Ayurveda, Siddha and Unani with WHO's ICD-11"
	if os.Getenv("RENDER") != "" {
		docs.SwaggerInfo.Host = host
	} else {
		docs.SwaggerInfo.Host = host + ":" + port
	}
	docs.SwaggerInfo.BasePath = "/api/v1"

	httpClient := http.Client{}

	// Get the ICD API IDs etc
	icdClientID := os.Getenv("ICD_CLIENTID")
	icdClientSecret := os.Getenv("ICD_CLIENTSECRET")

	genaiClient, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Set up the repositories
	icdRepository := repository.NewICDRepository(&httpClient, icdClientID, icdClientSecret)
	namasteRepository := repository.NewNamasteRepository()

	// Set up services
	autocompleteService := service.NewAutoComplete(genaiClient, icdRepository, namasteRepository)
	codeSystemService := service.NewCodeSystemService(namasteRepository, icdRepository)

	// Set up controllers
	autocompleteController := controller.NewAutocompleteController(autocompleteService)
	databaseController := controller.NewDatabaseController(autocompleteService)
	serverController := controller.NewServerController()
	codeSystemController := controller.NewCodeSystemController(codeSystemService)

	// Rate limiter
	rate, err := limiter.NewRateFromFormatted("10-M")
	if err != nil {
		log.Fatalln("Failed to create rate limiter: %w", err)
	}

	// Setup middlewares
	rateLimitStore := memory.NewStore()
	rateLimiterMiddleware := mgin.NewMiddleware(limiter.New(rateLimitStore, rate))

	cacheStore := persistence.NewInMemoryStore(time.Hour)

	apiRoutes := r.Group(docs.SwaggerInfo.BasePath)
	apiRoutes.Use(rateLimiterMiddleware)
	{
		codeSystemRoutes := apiRoutes.Group("/codesystem")
		{
			codeSystemRoutes.GET("/namaste", cache.CachePage(cacheStore, time.Hour, codeSystemController.ListNamaste))
		}

		apiRoutes.GET("/sync", databaseController.Sync)
		apiRoutes.GET("/autocomplete", cache.CachePage(cacheStore, time.Hour, autocompleteController.Find))
		apiRoutes.GET("/health", serverController.Health)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":" + port)
}
