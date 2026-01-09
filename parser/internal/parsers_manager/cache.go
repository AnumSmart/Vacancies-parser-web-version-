package parsers_manager

import (
	"fmt"
	"parser/internal/domain/models"
	"strings"
)

// метод получения данных из поикового кэша по заданному хэшу поиска
func (pm *ParsersManager) tryGetFromCache(params models.SearchParams) ([]models.SearchVacanciesResult, bool) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("⏳ Ищем вакансии в кэше...")

	searchHash, err := pm.generateSearchHash(params)
	if err != nil {
		fmt.Printf("⚠️  Ошибка генерации поискового хэша: %v\n", err)
		return nil, false
	}

	// пытаемся найти в кэше данные по заданному хэш ключу
	cached, ok := pm.SearchCache.GetItem(searchHash)
	if !ok {
		fmt.Println("⏳ Не удалось найти данные в кэше!")
		return nil, false
	}

	// если можно получить данные из кэша №1, то получаем интерфейс.
	// проводим type assertion, проверяем нужный тип
	results, ok := cached.([]models.SearchVacanciesResult)
	if !ok {
		fmt.Println("⚠️  Type assertion для кэшированных данных не удался")
		return nil, false
	}

	fmt.Println("✅ Найдены кэшированные данные")
	return results, true
}

// метод для кэширования результатов поиска списка вакансий в 2 кэша (в поисковый кэш и в индексный)
func (pm *ParsersManager) cacheSearchResults(params models.SearchParams, results []models.SearchVacanciesResult) {
	searchHash, err := pm.generateSearchHash(params)
	if err != nil {
		fmt.Printf("⚠️  Не удалось кэшировать результаты: %v\n", err)
		return
	}

	//записываем данные в поисковый кэш №1
	pm.SearchCache.AddItemWithTTL(searchHash, results, pm.config.Cache.SearchCacheConfig.SearchCacheTTL)

	// Строим обратный индекс и сразу кэшируем его в кэше №2
	pm.buildReverseIndex(searchHash, results)

	fmt.Printf("✅ Результаты поиска закэшированы в поисковом кэше (ключ: %s)\n", searchHash)
}

// метод для кэширования результатов поиска деталей конкретной вакансии по заднанному ID и парсеру (источнику)
func (pm *ParsersManager) cacheDetailsResult(vacancyID string, results models.SearchVacancyDetailesResult) {
	//записываем данные в поисковый кэш №3 (для деталей вакансии)
	pm.SearchCache.AddItemWithTTL(vacancyID, results, pm.config.Cache.VacancyCacheConfig.VacancyCacheTTL)

	fmt.Printf("✅ Результаты поиска закэшированы в поисковом кэше (ключ: %s)\n", vacancyID)
}

// метод обёртка для генерации поискового хэша
func (pm *ParsersManager) generateSearchHash(params models.SearchParams) (string, error) {
	return genHashFromSearchParam(params)
}
