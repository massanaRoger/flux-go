package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flux/internal/adapters/database"
	"flux/internal/adapters/queue"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	jobRepo := database.NewPostgresJobRepository(db)
	queueBroker := queue.NewRedisQueueBroker()
	defer queueBroker.Close()

	// TODO: Initialize services and handlers
	_ = jobRepo
	_ = queueBroker

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "flux-api-server",
		})
	})

	// Metrics endpoint (Prometheus)
	router.GET("/metrics", func(c *gin.Context) {
		// TODO: Implement Prometheus metrics
		c.String(http.StatusOK, "# HELP flux_api_requests_total Total API requests\n")
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// TODO: Use proper handlers
		v1.POST("/jobs", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) })
		v1.GET("/jobs/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) })
		v1.GET("/jobs", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) })
		v1.DELETE("/jobs/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) })
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting Flux API Server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

