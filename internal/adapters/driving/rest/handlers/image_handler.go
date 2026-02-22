package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
	"github.com/llascola/web-backend/internal/app/outports"
)

func (h *Handler) UploadImage(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	url, err := h.imageService.UploadImage(ctx, file, outports.FileMetadata{
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.UploadImageResponse{Url: url})
}
