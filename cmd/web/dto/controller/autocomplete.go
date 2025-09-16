package controller

import (
	"backend/cmd/web/dto"
	"backend/internal/service"
	"net/http"

	_ "backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AutocompleteController interface {
	Find(ctx *gin.Context)
}

type autocompleteController struct {
	service service.AutoCompleteService
}

// @Summary		Retrive matches
// @Description	Retrieves matches by combining results from ICD and NAMASTE repositories
// @Produce		json
// @Param		query query string true "Search query"
// @Success		200		{object}	service.Matches
// @Failure		500		{object}	dto.Error
// @Router			/autocomplete [get]
func (a *autocompleteController) Find(ctx *gin.Context) {
	query := ctx.Query("query")

	resp, err := a.service.Find(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func NewAutocompleteController(service service.AutoCompleteService) AutocompleteController {
	return &autocompleteController{
		service: service,
	}
}
