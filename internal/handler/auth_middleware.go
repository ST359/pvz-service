package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authHeader = "Authorization"
	userRole   = "userRole"
)

func (h *Handler) userRoleMW(c *gin.Context) {
	const op = "handler.auth.userRole"

	header := c.GetHeader(authHeader)

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" || len(headerParts[1]) == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrMessageUnauthorized)
		return
	}

	role, err := h.services.User.ParseToken(headerParts[1])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrMessageUnauthorized)
		return
	}

	c.Set(userRole, role)
	c.Next()
}
