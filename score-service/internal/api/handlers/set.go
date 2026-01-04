package handlers

import (
	"archy/scores/internal/core/models"
	"archy/scores/internal/core/services"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type SetHandler struct {
	service *services.SetService
}

func NewSetHandler(service *services.SetService) *SetHandler {
	return &SetHandler{service: service}
}

func (h *SetHandler) RegisterRoutes(e *echo.Echo) {
	group := e.Group("/api/rounds/:roundId/sets")

	group.GET("", h.ListSets)
	group.POST("", h.CreateSet)
	group.GET("/:setId", h.GetSet)
	group.GET("/:setId/shots", h.GetSetShots)
}

func (h *SetHandler) CreateSet(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	var req models.CreateSetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if req.SetNumber <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Set number must be positive",
		})
	}

	set, err := h.service.CreateSet(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to create set",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, set)
}

func (h *SetHandler) ListSets(c echo.Context) error {
	externalUserID, ok := c.Get("external_user_id").(string)
	if !ok || externalUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	roundID := pgtype.UUID{}
	err := roundID.Scan(c.Param("roundId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid round ID format",
		})
	}

	sets, err := h.service.GetSetsForQualificationRound(c.Request().Context(), roundID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to fetch sets",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, sets)
}

func (h *SetHandler) GetSet(c echo.Context) error {
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

	set, err := h.service.GetSet(c.Request().Context(), setID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Set not found or access denied",
		})
	}

	return c.JSON(http.StatusOK, set)
}

func (h *SetHandler) GetSetShots(c echo.Context) error {
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

	shots, err := h.service.GetSetShots(c.Request().Context(), setID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to fetch shots",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, shots)
}
