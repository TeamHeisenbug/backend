package controller

import (
	"backend/cmd/web/dto"
	"backend/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CodeSystemController interface {
	ListNamaste(ctx *gin.Context)
}

type codeSystemController struct {
	codeSystemService service.CodeSystemService
}

// @Summary		List all namaste codes
// @Tags Code System
// @Param		size query int false "Number of codes you want"
// @Produce		json
// @Success		200		{object}	dto.CodeSystem
// @Failure		500		{object}	dto.Error
// @Router			/codesystem/namaste [get]
func (c *codeSystemController) ListNamaste(ctx *gin.Context) {
	var size int
	sizeQuery := ctx.Query("size")
	if sizeQuery == "" {
		size = 5000
	} else {
		var err error
		size, err = strconv.Atoi(sizeQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.Error{Error: fmt.Sprintf("unable to parse size: %w", err)})
			return
		}
	}

	// TODO: Stop hard coding this in future
	url := "https://backend-kl02.onrender.com/api/v1/codesystem/namaste"
	codeSystem, err := c.codeSystemService.ListNamaste(size, url)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, codeSystem)
}

func NewCodeSystemController(codeSystemService service.CodeSystemService) CodeSystemController {
	return &codeSystemController{
		codeSystemService: codeSystemService,
	}
}
