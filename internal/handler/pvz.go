package handler

import (
	"log/slog"
	"net/http"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePVZ(c *gin.Context) {
	const op = "handler.pvz.CreatePVZ"

	role, _ := c.Get(userRole)
	if role != api.UserRoleModerator {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrMessageAccessDenied)
		return
	}
	var pvzreq api.PostPvzJSONRequestBody
	err := c.ShouldBind(&pvzreq)
	if err != nil {
		h.Logger.Error("failed to bind pvz request", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}

	pvzres, err := h.Services.PVZ.Create(pvzreq)
	if err != nil {
		h.Logger.Error("failed to create pvz", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, pvzres)
}
func (h *Handler) GetPVZ(c *gin.Context) {
	const op = "handler.pvz.GetPVZ"
	//auth handled in middleware
	var params api.GetPvzParams

	err := c.ShouldBind(&params)
	if err != nil {
		h.Logger.Error("failed to bind pvz get request", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	info, err := h.Services.GetByDate(params)
	if err != nil {
		h.Logger.Error("failed to get pvz by date", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusOK, info)
}
