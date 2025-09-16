package controller

import (
	"backend/cmd/web/dto"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabaseController interface {
	Sync(ctx *gin.Context)
}

type databaseController struct {
	service service.AutoCompleteService
}

// @Summary		Syncs databases
// @Description	Updates the database index with the NAMASTE CSV files' data
// @Produce		json
// @Success		200		{object}	dto.Message
// @Failure		500		{object}	dto.Error
// @Router			/sync [get]
func (d *databaseController) Sync(ctx *gin.Context) {
	if err := d.service.Update(); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.Message{Message: "All repositories synced"})
}

func NewDatabaseController(autocompleteService service.AutoCompleteService) DatabaseController {
	return &databaseController{
		service: autocompleteService,
	}
}
