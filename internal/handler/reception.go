package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreateReception(c *gin.Context) {
	const op = "handler.reception.CreateReception"

	role, _ := c.Get(userRole)
	if role != api.UserRoleEmployee {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrMessageAccessDenied)
		return
	}
	var pvzID api.PostReceptionsJSONBody
	err := c.ShouldBind(&pvzID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	reception, err := h.Services.Reception.Create(pvzID.PvzId)
	if err != nil {
		if errors.Is(err, errs.ErrReceptionNotClosed) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.Logger.Error("failed to open reception", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, reception)
}
func (h *Handler) CloseLastReception(c *gin.Context) {
	const op = "handler.reception.CloseLastReception"

	pvzIdStr := c.Param("pvzId")

	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	reception, err := h.Services.Reception.CloseLastReception(pvzId)
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.Logger.Error("failed to close reception", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusOK, reception)
}
func (h *Handler) DeleteLastProduct(c *gin.Context) {
	const op = "handler.reception.DeleteLastProduct"

	pvzIdStr := c.Param("pvzId")

	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	err = h.Services.Reception.DeleteLastProduct(pvzId)
	if err != nil {
		if errors.Is(err, errs.ErrNoProductsInReception) || errors.Is(err, errs.ErrNoReceptionsInProgress) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.Logger.Error("failed to delete last product", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
func (h *Handler) AddProduct(c *gin.Context) {
	const op = "handler.reception.AddProduct"
	role, _ := c.Get(userRole)
	if role != api.UserRoleEmployee {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrMessageAccessDenied)
		return
	}
	var prodReq api.PostProductsJSONBody
	err := c.ShouldBind(&prodReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	prodRes, err := h.Services.AddProduct(prodReq.PvzId, api.ProductType(prodReq.Type))
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.Logger.Error("failed to add product", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, prodRes)
}
