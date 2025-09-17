package controller

import (
	"backend/cmd/web/dto"
	"backend/internal/service"
	"net/http"
	"time"

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
// @Success		200		{object}	[]dto.ValueSet
// @Failure		500		{object}	dto.Error
// @Router			/autocomplete [get]
func (a *autocompleteController) Find(ctx *gin.Context) {
	query := ctx.Query("query")

	resp, err := a.service.Find(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error{Error: err.Error()})
		return
	}

	valueSets := make([]dto.ValueSet, 0)
	for _, disease := range resp.Diseases {
		valueSets = append(valueSets, dto.ValueSet{
			ResourceType: "ValueSet",
			ID:           "autocomplete-results",
			Status:       "active",
			Expansion: dto.Expansion{
				Identifier: "https://backend-kl02.onrender.com/api/v1/autocomplete",
				Timestamp:  time.Now(),
				Total:      2,
				Offset:     0,
				Contains: []dto.Contain{
					{
						System:  "https://backend-kl02.onrender.com/api/v1/codesystem/namaste",
						Code:    disease.Namaste.ID,
						Display: disease.Namaste.Name,
						Extension: dto.Extension{
							URL:         "https://backend-kl02.onrender.com/api/v1/structuredefinition/sourceSystem",
							ValueString: "NAMASTE",
						},
					},
					{
						System:  "https://backend-kl02.onrender.com/api/v1/codesystem/icd",
						Code:    disease.ICD.ID,
						Display: disease.ICD.Name,
						Extension: dto.Extension{
							URL:         "https://backend-kl02.onrender.com/api/v1/structuredefinition/sourceSystem",
							ValueString: "ICD",
						},
					},
				},
			},
		})
	}

	ctx.JSON(http.StatusOK, valueSets)
}

func NewAutocompleteController(service service.AutoCompleteService) AutocompleteController {
	return &autocompleteController{
		service: service,
	}
}
