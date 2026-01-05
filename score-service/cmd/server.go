package main

import (
	"archy/scores/internal/api/handlers"
	"archy/scores/internal/core/services"
	"archy/scores/internal/db"
	"archy/scores/jwt"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

var connectionStr = "postgresql://archy:12345678@localhost:5430/archy?sslmode=disable"

func main() {

	dbpool, err := pgxpool.New(context.Background(), connectionStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbpool.Close()

	// Проверка соединения
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	e := echo.New()

	queries := db.New(dbpool)
	// Get auth service URL from environment variable
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:3000" // Default to local development
	}

	// Initialize JWT verifier
	jwtVerifier := jwt.NewJWKVerifier(authServiceURL)
	if err := jwtVerifier.Initialize(); err != nil {
		e.Logger.Fatal("Failed to initialize JWT verifier: ", err)
	}

	// Public endpoint (no auth required)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Archy!")
	})

	// Protected endpoint (requires valid JWT)
	protected := e.Group("/api")
	protected.Use(jwtVerifier.JWTMiddleware())
	//protected.GET("/score", func(c echo.Context) error {
	//	// Get user ID from context (set by middleware)
	//	userID := c.Get("user_id")
	//	return c.JSON(http.StatusOK, map[string]interface{}{
	//		"message": "Protected endpoint accessed successfully",
	//		"user_id": userID,
	//	})
	//})
	roundService := services.NewQualificationRoundService(queries)
	setService := services.NewSetService(queries)
	shotService := services.NewShotService(queries)

	shotHandler := handlers.NewShotHandler(shotService)
	setHandler := handlers.NewSetHandler(setService)
	roundHandler := handlers.NewQualificationRoundHandler(roundService)

	shotHandler.RegisterRoutes(e)
	setHandler.RegisterRoutes(e)
	roundHandler.RegisterRoutes(e)

	e.Logger.Fatal(e.Start(":1323"))
}
