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

	var role api.PostDummyLoginJSONBody
	err := c.ShouldBind(&role)
	if err != nil {
		h.Logger.Error("failed to bind dummy login request", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	if string(role.Role) != string(api.UserRoleEmployee) && string(role.Role) != string(api.UserRoleModerator) {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	tok, err := h.Services.GenerateToken(string(role.Role))
	if err != nil {
		h.Logger.Error("failed to generate dummy token", slog.String("op", op), slog.String("error", err.Error()))
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
		h.Logger.Error("failed to bind login credentials", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	tok, err := h.Services.Login(creds)
	if err != nil {
		if errors.Is(err, errs.ErrWrongCreds) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrMessageWrongCredentials)
			return
		}
		h.Logger.Error("failed to login user", slog.String("op", op), slog.String("error", err.Error()))
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
		h.Logger.Error("failed to bind register credentials", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
		return
	}
	user, err := h.Services.CreateUser(creds)
	if err != nil {
		if errors.Is(err, errs.ErrEmailExists) || errors.Is(err, errs.ErrPasswordTooLong) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrMessageBadRequest)
			return
		}
		h.Logger.Error("failed to register user", slog.String("op", op), slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrMessageInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, user)
}
