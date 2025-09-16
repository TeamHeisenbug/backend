package service

import (
	"backend/cmd/web/dto"
	"backend/internal/repository"
)

type CodeSystemService interface {
	ListNamaste(size int, url string) (*dto.CodeSystem, error)
	ListICD() (*dto.CodeSystem, error)
}

type codeSystemService struct {
	namasteRepository repository.NamasteRepository
	icdRepository     repository.ICDRepository
}

// ListICD implements CodeSystemService.
func (c *codeSystemService) ListICD() (*dto.CodeSystem, error) {
	panic("unimplemented")
}

// ListNamaste implements CodeSystemService.
func (c *codeSystemService) ListNamaste(size int, url string) (*dto.CodeSystem, error) {
	list, err := c.namasteRepository.List(size)
	if err != nil {
		return nil, err
	}

	var result dto.CodeSystem

	result.ResourceType = "CodeSystem"
	result.Version = "1.0"
	result.Status = "active"
	result.Content = "complete"
	result.ID = "NAMASTE"
	result.Name = "NAMASTE Codes"
	result.URL = url
	result.Concept = make([]dto.Concept, 0)

	for _, match := range list {
		result.Concept = append(result.Concept, dto.Concept{
			Code:       match.ID,
			Display:    match.Name,
			Definition: match.Desc,
			Property: []dto.Property{
				{
					Code:        "type",
					ValueString: match.Type,
				},
			},
		})
	}

	return &result, nil
}

func NewCodeSystemService(namasteRepository repository.NamasteRepository, icdRepository repository.ICDRepository) CodeSystemService {
	return &codeSystemService{
		namasteRepository: namasteRepository,
		icdRepository:     icdRepository,
	}
}
