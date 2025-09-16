package controller

import (
	"backend/cmd/web/dto"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServerController interface {
	Health(ctx *gin.Context)
}

type serverController struct {
	service service.AutoCompleteService
}

// @Summary		Check if server is alive
// @Produce		json
// @Success		200		{object}	dto.Message
// @Router			/health [get]
func (d *serverController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, dto.Message{Message: "ok"})
}

func NewServerController() ServerController {
	return &serverController{}
}
