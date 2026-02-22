package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
)

func (h *Handler) Register(ctx *gin.Context) {
	var req openapi.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Register(ctx, string(req.Email), req.Password); err != nil {
		HandleError(ctx, err)
		return
	}

	msg := "User registered successfully"
	ctx.JSON(http.StatusCreated, openapi.MessageResponse{Message: &msg})
}

func (h *Handler) Login(ctx *gin.Context) {
	var req openapi.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(ctx, string(req.Email), req.Password)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Logout(ctx *gin.Context) {
	jtiRaw, exists := ctx.Get("jti")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing jti tracking in context"})
		return
	}
	expRaw, existsExp := ctx.Get("exp")
	if !existsExp {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing exp tracking in context"})
		return
	}

	jti, ok1 := jtiRaw.(string)
	expFloat, ok2 := expRaw.(float64)
	if !ok1 || !ok2 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token tracking claims"})
		return
	}

	expiration := time.Unix(int64(expFloat), 0)

	if err := h.authService.Logout(ctx.Request.Context(), jti, expiration); err != nil {
		HandleError(ctx, err)
		return
	}

	msg := "Successfully logged out"
	ctx.JSON(http.StatusOK, openapi.MessageResponse{Message: &msg})
}

func (h *Handler) Refresh(ctx *gin.Context) {
	var req openapi.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
