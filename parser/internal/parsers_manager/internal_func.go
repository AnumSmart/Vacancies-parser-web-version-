package parsers_manager

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"parser/internal/domain/models"
)

// Метод для вывода в консоль результатов поиска списка вакансий (с нужными атрибутами)
func (pm *ParsersManager) printMultiSearchResults(results []models.SearchVacanciesResult, resultsPerPage int) {
	totalVacancies := 0

	for _, result := range results {
		fmt.Printf("\n📊 %s:\n", result.ParserName)
		fmt.Printf("   ⏱️  Время: %v\n", result.Duration)

		if result.Error != nil {
			fmt.Printf("   ❌ Ошибка: %v\n", result.Error)
			continue
		}

		fmt.Printf("   ✅ Найдено: %d вакансий\n", len(result.Vacancies))
		totalVacancies += len(result.Vacancies)

		// Показываем столько результатов, сколько ввели, или если их меньше --- столько, сколько нашли
		for i, vacancy := range result.Vacancies {
			if i >= resultsPerPage {
				break
			}
			fmt.Printf("      %d. %s - %s, company:%s, URL:[ %s ], ID:%s\n", i+1, vacancy.Job, *vacancy.Salary, vacancy.Company, vacancy.URL, vacancy.ID)
		}

		if len(result.Vacancies) > resultsPerPage {
			fmt.Printf("      ... и ещё %d\n", len(result.Vacancies)-resultsPerPage)
		}
	}

	fmt.Printf("\n🎯 Всего найдено: %d вакансий\n", totalVacancies)
}

// метод для построения обратного индекса и хранения его в кэше №2 для индексов и ID вакансий
func (pm *ParsersManager) buildReverseIndex(searchHash string, results []models.SearchVacanciesResult) {
	for _, parserResult := range results {
		for i, vacancy := range parserResult.Vacancies {
			compositeID := fmt.Sprintf("%s_%s", vacancy.Source, vacancy.ID)

			indexEntry := models.VacancyIndex{
				SearchHash: searchHash,
				ParserName: parserResult.ParserName,
				Index:      i,
			}

			// Сохраняем в индексный кэш (ТОТ ЖЕ ТИП!), TTL такой же как для кэша поиска
			pm.VacancyIndex.AddItemWithTTL(compositeID, indexEntry, pm.config.Cache.VacancyCacheConfig.VacancyCacheTTL)
		}
	}
}

// функция генерирует хэш запроса поиска, чтобы кэшировать запросы по этому хэшу
func genHashFromSearchParam(params models.SearchParams) (string, error) {
	// Учитываем ВСЕ параметры, которые влияют на результат
	keyData := struct {
		Text    string `json:"text"`
		Area    string `json:"area"`
		PerPage int    `json:"per_page"`
		Page    int    `json:"page"`
		// Добавьте другие поля из SearchParams
	}{
		Text:    params.Text,
		Area:    params.Country,
		PerPage: params.PerPage,
		Page:    params.Page,
	}

	data, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("Error while marshaling of params: %w\n", err)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s", hex.EncodeToString(hash[:16])), nil
}
