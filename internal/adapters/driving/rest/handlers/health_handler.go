package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
