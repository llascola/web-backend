package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
)

func (h *Handler) Register(ctx *gin.Context) {
	var req openapi.RegisterJSONBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Register(ctx, string(req.Email), req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *Handler) Login(ctx *gin.Context) {
	var req openapi.LoginJSONBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(ctx, string(req.Email), req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
