package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/gin-gonic/gin"
)

func (h *Handler) DummyLogin(c *gin.Context) {
	const op = "handler.user.DummyLogin"

	var role string
	err := c.ShouldBind(&role)
	if err != nil {
		h.logger.Error("failed to bind dummy login request", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, ErrMessageBadRequest)
	}
	if role != string(api.UserRoleEmployee) && role != string(api.UserRoleModerator) {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	tok, err := h.services.GenerateToken(role)
	if err != nil {
		h.logger.Error("failed to generate dummy token", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusOK, api.Token(tok))
}
func (h *Handler) Login(c *gin.Context) {
	const op = "handler.user.Login"

	var creds api.PostLoginJSONBody

	err := c.ShouldBind(&creds)
	if err != nil {
		h.logger.Error("failed to bind login credentials", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, ErrMessageBadRequest)
	}
	tok, err := h.services.Login(creds)
	if err != nil {
		if errors.Is(err, errs.ErrWrongCreds) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrMessageWrongCredentials)
			return
		}
		h.logger.Error("failed to login user", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusOK, api.Token(tok))
}
func (h *Handler) Register(c *gin.Context) {
	const op = "handler.user.Register"

	var creds api.PostRegisterJSONBody

	err := c.ShouldBind(&creds)
	if err != nil {
		h.logger.Error("failed to bind register credentials", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, ErrMessageBadRequest)
	}
	user, err := h.services.CreateUser(creds)
	if err != nil {
		if errors.Is(err, errs.ErrEmailExists) || errors.Is(err, errs.ErrPasswordTooLong) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.logger.Error("failed to register user", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, user)
}
