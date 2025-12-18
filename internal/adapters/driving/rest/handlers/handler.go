package handlers

import (
	"github.com/llascola/web-backend/internal/app"
	"github.com/llascola/web-backend/internal/app/inports"
)

type Handler struct {
	authService  inports.AuthService
	imageService inports.ImageService
	userService  inports.UserService
}

func NewHandler(app *app.Application) *Handler {
	return &Handler{
		authService:  app.Service.AuthService,
		imageService: app.Service.ImageService,
		userService:  app.Service.UserService,
	}
}
