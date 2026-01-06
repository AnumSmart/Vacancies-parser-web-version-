package converters

import (
	"parser/internal/domain/models"
	"parser/internal/search_server/dto"
)

// конвертация данных из DTO в доменную область
// в параменте передаём значение, чтобы не модифицировать входные данные
func SearchRequestToParams(req dto.SearchRequest) models.SearchParams {
	return models.SearchParams{
		Text:     req.Query,
		Location: req.Location,
		PerPage:  req.PerPage,
		Page:     req.Page,
	}
}
