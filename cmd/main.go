package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expertlisting/internal/cache"
	"expertlisting/internal/config"
	"expertlisting/internal/database"
	"expertlisting/internal/handlers"
	"expertlisting/internal/middleware"
	"expertlisting/internal/repository"
	"expertlisting/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "expertlisting/docs"
)

// @title ExpertListing Property Management API
// @version 1.0
// @description Property Listings API with proximity search, filtering, and pagination.
// @BasePath /api
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Using system environment variables")
	}

	cfg := config.LoadConfig()

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	database.SeedDemoData(db)

	memCache := cache.NewMemoryCache()
	listingRepo := repository.NewListingRepository(db)
	listingService := services.NewListingService(listingRepo, memCache)
	listingHandler := handlers.NewListingHandler(listingService)
	healthHandler := handlers.NewHealthHandler()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.MaxBodySize(1 << 20))
	router.Use(middleware.RateLimiter(120, time.Minute/120))
	router.Use(middleware.RequestLogger())
	router.Use(middleware.PanicRecovery())
	router.Use(middleware.SetupCORS(cfg.AllowOrigins))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/health", healthHandler.Check)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(1),
	))

	registerRoutes := func(rg *gin.RouterGroup) {
		listings := rg.Group("/listings")
		{
			listings.GET("/search", listingHandler.SearchListings)
			listings.GET("", listingHandler.GetListings)
			listings.POST("", listingHandler.CreateListing)
			listings.GET("/:id", listingHandler.GetListingByID)
			listings.PUT("/:id", listingHandler.UpdateListing)
			listings.PATCH("/:id", listingHandler.PatchListing)
			listings.DELETE("/:id", listingHandler.DeleteListing)
		}
	}

	registerRoutes(router.Group("/api"))
	registerRoutes(router.Group("/api/v1"))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
