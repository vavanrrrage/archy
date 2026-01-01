package main

import (
	"archy/scores/jwt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

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
	protected.GET("/score", func(c echo.Context) error {
		// Get user ID from context (set by middleware)
		userID := c.Get("user_id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Protected endpoint accessed successfully",
			"user_id": userID,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
