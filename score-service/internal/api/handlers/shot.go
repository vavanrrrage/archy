package handlers

import (
	"archy/scores/internal/core/models"
	"archy/scores/internal/core/services"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ShotHandler struct {
	service *services.ShotService
}

func NewShotHandler(service *services.ShotService) *ShotHandler {
	return &ShotHandler{service: service}
}

func (h *ShotHandler) RegisterRoutes(e *echo.Echo) {
	group := e.Group("/api/rounds/:roundId/sets/:setId/shots")

	group.GET("", h.ListShots)
	group.POST("", h.CreateShot)
	//group.POST("/batch", h.CreateShotsBatch)
	group.GET("/:shotId", h.GetShot)
}

func (h *ShotHandler) CreateShot(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	setID, err := uuid.Parse(c.Param("setId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid set ID format",
		})
	}

	var req models.CreateShotRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Валидация координат
	if req.X == 0 && req.Y == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Coordinates cannot be both zero",
		})
	}

	shot, err := h.service.CreateShot(
		c.Request().Context(),
		setID,
		req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to create shot",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, shot)
}

//// CreateShotsBatch создает несколько выстрелов за раз
//// POST /api/rounds/:roundId/sets/:setId/shots/batch
//func (h *ShotHandler) CreateShotsBatch(c echo.Context) error {
//	externalUserID, ok := c.Get("external_user_id").(string)
//	if !ok || externalUserID == "" {
//		return c.JSON(http.StatusUnauthorized, map[string]string{
//			"error": "User not authenticated",
//		})
//	}
//
//	roundID, err := uuid.Parse(c.Param("roundId"))
//	if err != nil {
//		return c.JSON(http.StatusBadRequest, map[string]string{
//			"error": "Invalid round ID format",
//		})
//	}
//
//	setID, err := uuid.Parse(c.Param("setId"))
//	if err != nil {
//		return c.JSON(http.StatusBadRequest, map[string]string{
//			"error": "Invalid set ID format",
//		})
//	}
//
//	var req models.CreateShotsBatchRequest
//	if err := c.Bind(&req); err != nil {
//		return c.JSON(http.StatusBadRequest, map[string]string{
//			"error": "Invalid request format",
//		})
//	}
//
//	// Валидация
//	if len(req.Shots) == 0 {
//		return c.JSON(http.StatusBadRequest, map[string]string{
//			"error": "No shots provided",
//		})
//	}
//
//	if len(req.Shots) > 12 {
//		return c.JSON(http.StatusBadRequest, map[string]string{
//			"error": "Maximum 12 shots per batch",
//		})
//	}
//
//	shots, err := h.service.CreateShotsBatch(c.Request().Context(), externalUserID, roundID, setID, req)
//	if err != nil {
//		return c.JSON(http.StatusInternalServerError, map[string]string{
//			"error":   "Failed to create shots",
//			"details": err.Error(),
//		})
//	}
//
//	return c.JSON(http.StatusCreated, shots)
//}

func (h *ShotHandler) ListShots(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	setID, err := uuid.Parse(c.Param("setId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid set ID format",
		})
	}

	shots, err := h.service.GetShotsBySet(c.Request().Context(), setID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to fetch shots",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, shots)
}

func (h *ShotHandler) GetShot(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	shotID, err := uuid.Parse(c.Param("shotId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid shot ID format",
		})
	}

	shot, err := h.service.GetShot(c.Request().Context(), shotID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Shot not found or access denied",
		})
	}

	return c.JSON(http.StatusOK, shot)
}
