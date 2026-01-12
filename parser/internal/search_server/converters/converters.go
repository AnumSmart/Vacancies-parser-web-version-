package converters

import (
	"parser/internal/domain/models"
	"parser/internal/search_server/dto"
	"strconv"
)

// конвертация данных из DTO в доменную область
// в параменте передаём значение, чтобы не модифицировать входные данные
func SearchRequestDTOToParamsDomain(req dto.SearchRequest) models.SearchParams {
	return models.SearchParams{
		Text:     req.Query,
		Location: req.Location,
		PerPage:  req.PerPage,
		Page:     req.Page,
	}
}

// конвертация данных для DTO слоя
// разделение найденных данных по источникам
func SearchVacanciesResultDomainToDTO(domainResults []models.SearchVacanciesResult) dto.SearchVacanciesResponse {
	response := dto.SearchVacanciesResponse{
		Results: make(map[string]dto.SourceVacancies),
		Total:   0,
	}

	for _, domainResult := range domainResults {
		sourceKey := domainResult.ParserName // "hh", "superjob", etc.

		// Создаем или получаем запись для этого источника
		sourceVacancies, exists := response.Results[sourceKey]
		if !exists {
			sourceVacancies = dto.SourceVacancies{
				Name:  getSourceName(domainResult.ParserName),
				Icon:  getSourceIcon(domainResult.ParserName),
				Count: 0,
			}
		}

		// Обрабатываем ошибку
		if domainResult.Error != nil {
			sourceVacancies.HasError = true
			sourceVacancies.Error = domainResult.Error.Error()
		} else {
			// Конвертируем вакансии
			vacanciesDTO := make([]dto.VacancyResponse, 0, len(domainResult.Vacancies))
			for _, vacancy := range domainResult.Vacancies {
				vacancyDTO := ConvertVacancyToDTO(vacancy)
				vacanciesDTO = append(vacanciesDTO, vacancyDTO)
			}

			sourceVacancies.Vacancies = append(sourceVacancies.Vacancies, vacanciesDTO...)
			sourceVacancies.Count = len(sourceVacancies.Vacancies)
		}

		// Добавляем информацию о времени выполнения
		if domainResult.Duration > 0 {
			sourceVacancies.Duration = formatDuration(domainResult.Duration)
		}

		// Обновляем в мапе
		response.Results[sourceKey] = sourceVacancies

		// Считаем общее количество
		response.Total += sourceVacancies.Count
	}

	return response
}

// Вспомогательная функция для конвертации одной вакансии
func ConvertVacancyToDTO(vacancy models.Vacancy) dto.VacancyResponse {
	dtoVacancy := dto.VacancyResponse{
		ID:          vacancy.ID,
		Job:         vacancy.Job,
		Company:     vacancy.Company,
		Salary:      formatSalary(vacancy.Salary, vacancy.Currency),
		Currency:    vacancy.Currency,
		Location:    vacancy.Location,
		Experience:  formatExperience(vacancy.Experience),
		Schedule:    formatSchedule(vacancy.Schedule),
		URL:         vacancy.URL,
		Description: vacancy.Description,
		PublishedAt: formatPublishedAt(vacancy.PublishedAt),
	}

	// Заполняем информацию об источнике
	dtoVacancy.Source.Name = getSourceName(vacancy.Source)
	dtoVacancy.Source.Icon = getSourceIcon(vacancy.Source)

	return dtoVacancy
}

// Функция для конвертации информации по вакансии с описанием
func ConvertVacancyResultInfoDomainToDTO(resp models.SearchVacancyDetailesResult) dto.VacancyDetailsResponce {
	dtoVacancyInfo := dto.VacancyDetailsResponce{
		ID:          resp.ID,
		Job:         resp.Name,
		Company:     resp.Employer.Name,
		Salary:      strconv.Itoa(resp.Salary.From),
		Description: resp.Description,
		URL:         resp.Url,
	}

	return dtoVacancyInfo
}
