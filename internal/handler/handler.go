package handler

import (
	"log/slog"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	ErrMessageAccessDenied        = api.Error{Message: "Access denied"}
	ErrMessageBadRequest          = api.Error{Message: "Bad request"}
	ErrMessageInternalServerError = api.Error{Message: "Internal server error"}
	ErrMessageWrongCredentials    = api.Error{Message: "Wrong credentials"}
)

type Handler struct {
	Services *service.Service
	Logger   *slog.Logger
}

func NewHandler(service *service.Service, Logger *slog.Logger) *Handler {
	return &Handler{Services: service, Logger: Logger}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	public := r.Group("/")
	{
		public.POST("/dummyLogin", h.DummyLogin)
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
	}

	// Routes with auth
	protected := r.Group("/")
	protected.Use(h.userRoleMW)
	{
		protected.POST("/pvz", h.CreatePVZ)
		protected.GET("/pvz", h.GetPVZ)
		protected.POST("/pvz/:pvzId/close_last_reception", h.CloseLastReception)
		protected.POST("/pvz/:pvzId/delete_last_product", h.DeleteLastProduct)

		protected.POST("/receptions", h.CreateReception)

		protected.POST("/products", h.AddProduct)
	}
	return r
}
