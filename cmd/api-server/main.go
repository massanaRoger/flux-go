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
	httpAdapter "flux/internal/adapters/http"
	"flux/internal/adapters/queue"
	"flux/internal/app"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	jobRepo := database.NewPostgresJobRepository(dbPool)
	queueBroker := queue.NewRedisQueueBroker(redisClient)
	defer queueBroker.Close()

	// Initialize application services
	jobService := app.NewJobService(jobRepo)

	// Initialize handlers
	jobHandler := httpAdapter.NewJobHandler(jobService)

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
		v1.POST("/jobs", jobHandler.CreateJob)
		v1.GET("/jobs/:id", jobHandler.GetJob)
		v1.GET("/jobs", jobHandler.ListJobs)
		v1.DELETE("/jobs/:id", jobHandler.DeleteJob)
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
