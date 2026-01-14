package parsers_manager

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"search_service/internal/domain/models"
	"strconv"
	"strings"
)

// Главный метод (точка входа) логики поиска списка вакансий в зарегестрированных и "живых" парсерах
func (pm *ParsersManager) MultiSearch(scanner *bufio.Scanner) error {
	params := pm.getParamsForCLI(scanner)

	ctx := context.Background()

	// запускаем комплексный метод поиска
	results, err := pm.SearchVacancies(ctx, params)
	if err != nil {
		return err
	}

	// делаем несколько проверок. Проверка на nil результат, проверка на пустой слайс
	switch {
	case results == nil:
		log.Println("Внимание: получен nil")
		return fmt.Errorf("ошибка данных")
	case len(results) == 0:
		log.Println("Поиск не дал результатов")
		// Возможно, стоит возвращать специальную ошибку
		return fmt.Errorf("поиск не дал результатов")
	default:
		// вызываем функцию вывода в консоль информации о результатах поиска
		pm.printMultiSearchResults(results, params.PerPage)
	}

	return nil
}

// метод получения параметров запроса из командной строки для CLI реализации
func (pm *ParsersManager) getParamsForCLI(scanner *bufio.Scanner) models.SearchParams {
	fmt.Println("\n🌐 Мульти-поиск вакансий")

	var params models.SearchParams

	// читаем запрос пользователя и сохраняем его в параметрах
	fmt.Print("Введите поисковый запрос: ")
	if scanner.Scan() {
		params.Text = strings.TrimSpace(scanner.Text())
	}

	// читаем необходимое количество вакансий в поиске, но не более 50, сохраняем в параметрах
	fmt.Print("Количество вакансий на источник (max 50): ")
	if scanner.Scan() {
		countStr := strings.TrimSpace(scanner.Text())
		if countStr != "" {
			if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
				params.PerPage = count
			}
		}
	}

	if params.PerPage == 0 {
		params.PerPage = 20
	}

	return params
}
