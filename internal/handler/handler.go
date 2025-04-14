package handler

import (
	"log/slog"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	ErrMessageUnauthorized        = api.Error{Message: "Access denied"}
	ErrMessageBadRequest          = api.Error{Message: "Bad request"}
	ErrMessageInternalServerError = api.Error{Message: "Internal server error"}
	ErrMessageWrongCredentials    = api.Error{Message: "Wrong credentials"}
)

type Handler struct {
	services *service.Service
	logger   *slog.Logger
}

func NewHandler(service *service.Service, logger *slog.Logger) *Handler {
	return &Handler{services: service, logger: logger}
}
func (h *Handler) InitRoutes() *gin.Engine {
	//router := gin.New()

	panic("not implemented")
}
