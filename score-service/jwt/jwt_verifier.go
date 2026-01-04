package jwt

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type JWKVerifier struct {
	authServiceURL  string
	jwksURL         string
	Set             jwk.Set
	keysExpiry      time.Time
	mu              sync.RWMutex
	refreshInterval time.Duration
}

func NewJWKVerifier(authServiceURL string) *JWKVerifier {
	return &JWKVerifier{
		authServiceURL:  authServiceURL,
		jwksURL:         fmt.Sprintf("%s/api/auth/jwks", authServiceURL),
		refreshInterval: 15 * time.Minute,
	}
}

func (j *JWKVerifier) fetchJWKS() error {
	set, err := jwk.Fetch(context.Background(), j.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	j.Set = set
	j.keysExpiry = time.Now().Add(j.refreshInterval)
	return nil
}

func (j *JWKVerifier) getJWKSet() (jwk.Set, error) {
	j.mu.RLock()
	set := j.Set
	expiry := j.keysExpiry
	j.mu.RUnlock()

	if set != nil && time.Now().Before(expiry) {
		return set, nil
	}

	if err := j.fetchJWKS(); err != nil {
		return nil, err
	}

	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Set, nil
}

func (j *JWKVerifier) VerifyToken(tokenString string) (jwt.Token, error) {
	set, err := j.getJWKSet()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWK set: %w", err)
	}

	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(set),
		jwt.WithValidate(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return token, nil
}

func (j *JWKVerifier) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			const bearerPrefix = "Bearer "
			if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			tokenString := authHeader[len(bearerPrefix):]

			token, err := j.VerifyToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			}

			c.Set("token", token)

			if sub := token.Subject(); sub != "" {
				c.Set("external_user_id", sub)
			}

			return next(c)
		}
	}
}

func (j *JWKVerifier) Initialize() error {
	return j.fetchJWKS()
}
