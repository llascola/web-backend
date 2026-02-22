package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
)

func (h *Handler) HealthCheck(ctx *gin.Context) {
	status := "ok"
	now := time.Now()
	ctx.JSON(http.StatusOK, openapi.HealthResponse{
		Status:    &status,
		Timestamp: &now,
	})
}
