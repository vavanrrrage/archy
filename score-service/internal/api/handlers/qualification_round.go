package handlers

import (
	"archy/scores/internal/core/models"
	"archy/scores/internal/core/services"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type QualificationRoundHandler struct {
	service *services.QualificationRoundService
}

func NewQualificationRoundHandler(service *services.QualificationRoundService) *QualificationRoundHandler {
	return &QualificationRoundHandler{service: service}
}

func (h *QualificationRoundHandler) RegisterRoutes(e *echo.Echo) {
	group := e.Group("/api/rounds")

	group.GET("", h.GetUserRounds)
	group.POST("", h.CreateRound)
	group.GET("/:id", h.GetRound)
}

func (h *QualificationRoundHandler) CreateRound(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	var req models.CreateQualificationRoundRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	// Валидация
	if req.RoundType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Round type is required",
		})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Round name is required",
		})
	}
	if req.Distance <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Distance must be positive",
		})
	}
	if req.TotalSets <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Total sets must be positive",
		})
	}
	if req.ShotsPerSet <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Shots per set must be positive",
		})
	}

	// Устанавливаем время начала
	now := time.Now()
	req.StartTime = &now

	round, err := h.service.CreateQualificationRound(c.Request().Context(), externalUserID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to create round",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, round)
}

func (h *QualificationRoundHandler) GetUserRounds(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	rounds, err := h.service.GetQualificationRoundsForUser(c.Request().Context(), externalUserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to fetch rounds",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rounds)
}

func (h *QualificationRoundHandler) GetRound(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	roundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid round ID format",
		})
	}

	round, err := h.service.GetQualificationRound(c.Request().Context(), externalUserID, roundID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Round not found or access denied",
		})
	}

	return c.JSON(http.StatusOK, round)
}
