package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
)

func (h *Handler) GetProfile(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetProfile(ctx, userID)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.UserProfile{
		Id:    user.ID,
		Email: openapi_types.Email(user.Email),
		Role:  string(user.Role),
	})
}

func (h *Handler) DeleteUser(ctx *gin.Context, id openapi_types.UUID) {
	if err := h.userService.DeleteUser(ctx, id); err != nil {
		HandleError(ctx, err)
		return
	}

	msg := "User deleted"
	ctx.JSON(http.StatusOK, openapi.MessageResponse{Message: &msg})
}
