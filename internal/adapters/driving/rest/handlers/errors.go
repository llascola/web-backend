package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/app/domain"
)

// HandleError is the centralized HTTP Error translator. It examines pure Domain
// errors and maps them to standard HTTP status codes, outputting them in the
// standardized API response format.
func HandleError(ctx *gin.Context, err error) {
	var errValidation *domain.ErrValidation
	var errNotFound *domain.ErrNotFound
	var errConflict *domain.ErrConflict
	var errUnauthorized *domain.ErrUnauthorized
	var errInternal *domain.ErrInternal

	if errors.As(err, &errValidation) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errValidation.Message})
		return
	}

	if errors.As(err, &errNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": errNotFound.Message})
		return
	}

	if errors.As(err, &errConflict) {
		ctx.JSON(http.StatusConflict, gin.H{"error": errConflict.Message})
		return
	}

	if errors.As(err, &errUnauthorized) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errUnauthorized.Message})
		return
	}

	if errors.As(err, &errInternal) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInternal.Message})
		return
	}

	// Fallback for any unknown error types that slipped through
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
}
