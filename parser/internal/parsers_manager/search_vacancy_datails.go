package parsers_manager

import (
	"bufio"
	"context"
	"fmt"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"strconv"
	"strings"
	"time"
)

/*
// метод получения информации о вакансии из кэша с помощью кэша обратного индекса
func (pm *ParsersManager) GetVacancyDetails(scanner *bufio.Scanner) error {
	fmt.Println("\n📄 Детали вакансии (кратко):")

	// получаем ID вакансии и имя источника из ввода
	source, vacancyID, err := pm.getCompositeIDFromInput(scanner)
	if err != nil {
		return err
	}

	// создаём составной индекс, в котором будет ID вакансии и сервис, в котором этот ID нужно будет искать
	// этот составной индекс - будет ключем для кэша №2
	compositeID := fmt.Sprintf("%s_%s", source, vacancyID)

	// создаём переменную для искомой вакансии
	var targetVacancy models.Vacancy

	fmt.Println("⏳ Загружаем информацию...")

	// -------------------------------------------------------------------
	// пытаемся найти в кэше №2 данные по заданному ключу (составному индексу)
	searchResIndex, ok := pm.VacancyIndex.GetItem(compositeID)
	if !ok {
		return fmt.Errorf("No Vacancy with ID:%s was found in cache\n", vacancyID)
	}

	// проводим type assertion, проверяем нужный тип (так как нам функция GetItem возвращает интерфейс)
	searchResIndexChecked, ok := searchResIndex.(models.VacancyIndex)
	if !ok {
		fmt.Println("Type assertion after GetVacancyDetails ---> failed!")
		return fmt.Errorf("Type assertion after GetVacancyDetails ---> failed!\n")
	}

	// теперь из полученного из кэша индексов индекса мы можем найти нужный хэш запроса,
	// чтобы потом по этому хэшу из кэша поиска найти нужную вакансию по ID

	// пытаемся найти в кэше данные по заданному хэш ключу
	searchRes, ok := pm.SearchCache.GetItem(searchResIndexChecked.SearchHash)
	if ok {
		// если можно получить данные из кэша, то получаем интерфейс.
		// проводим type assertion, проверяем нужный тип
		searchResChecked, ok := searchRes.([]models.SearchVacanciesResult)
		if !ok {
			return fmt.Errorf("Type assertion after multi-search ---> failed!\n")
		}

		for _, neededElementRes := range searchResChecked {
			if neededElementRes.ParserName == source {
				for _, vacancyRes := range neededElementRes.Vacancies {
					if vacancyRes.ID == vacancyID {
						targetVacancy.ID = vacancyRes.ID
						targetVacancy.Job = vacancyRes.Job
						targetVacancy.Salary = vacancyRes.Salary
						targetVacancy.Company = vacancyRes.Company
						targetVacancy.Location = vacancyRes.Location
						targetVacancy.URL = vacancyRes.URL
					}
				}
			}
		}
	} else {
		pm.VacancyIndex.DeleteItem(compositeID)
		return fmt.Errorf("Данные устарели, сделайте повторный запрос (пункт меню 1)\n")
	}

	printVacancyDetails(targetVacancy)

	return nil
}

*/
// метод для получения полной информации по отдельной вакансии по ID
func (pm *ParsersManager) GetFullVacancyDetails(scanner *bufio.Scanner) error {
	// получаем ID вакансии и имя источника из ввода
	source, vacancyID, err := pm.getCompositeIDFromInput(scanner)
	if err != nil {
		return err
	}

	ctx := context.Background()

	result, err := pm.executeSearchVacancyDetailes(ctx, vacancyID, source)
	if err != nil {
		return err
	}

	// делаем несколько проверок. Проверка на nil результат, проверка на пустой слайс

	// создаём переменную для искомой вакансии
	var targetVacancy models.Vacancy

	salary := strconv.Itoa(result.Salary.From) // переводим зарплату из int в string

	targetVacancy.Company = result.Employer.Name
	targetVacancy.Job = result.Name
	targetVacancy.Description = result.Description
	targetVacancy.Salary = &salary
	targetVacancy.Location = result.Location.Name
	targetVacancy.ID = result.ID
	targetVacancy.URL = result.Url

	printVacancyDetails(targetVacancy)
	return nil
}

// метод менджера парсеров, который формирует джобу для поиска деталей по конкретной вакансии, добавляет эту джобу в очередь и получает результат поиска в канал
// возвращает результат поиска или ошибку
func (pm *ParsersManager) executeSearchVacancyDetailes(ctx context.Context, vacancyID, source string) (models.SearchVacancyDetailesResult, error) {
	// создаём новую джобу необходимого типа (в данном случае джоба поиска расширенной инфы по конкретной вакансии)
	job := pm.NewFetchVacancyJob(source, vacancyID)

	// Пытаемся добавить в очередь с таймаутом и повторными попытками
	success := pm.tryEnqueueJob(ctx, job, 5*time.Second)

	// проверяем успешность добавления в очередь
	if !success {
		return models.SearchVacancyDetailesResult{}, fmt.Errorf("❌ Джоба не была добавлена в очередь")
	}

	// дожидаемся результатов из очереди с учётом таймаута
	result, err := pm.waitForJobSearchVacancyDeyailsResult(ctx, job.ResultChan, 30*time.Second)

	// специально тут не обрабатываем ошибку, они уже обработаны выше
	return result, err
}

// Основная логика поиска деталей конкретной вакансии
func (pm *ParsersManager) searchVacancyDetailes(ctx context.Context, vacancyID, source string) (models.SearchVacancyDetailesResult, error) {
	// Проверяем кэш деталей вакансии
	// пытаемся найти в кэше данные по заданному хэш ключу
	cached, found := pm.vacancyDetails.GetItem(vacancyID)
	if found {
		// необходим type assertion
		checkedCached, ok := cached.(models.SearchVacancyDetailesResult)
		if !ok {
			return models.SearchVacancyDetailesResult{}, fmt.Errorf("⚠️  Type assertion для кэшированных данных деталей вакансии -  не удался\n")
		}
		return checkedCached, nil
	}

	// делаем проверку того, что источник(парсер) находтся в "живом состоянии"
	// согласно менеджеру статусов парсверов
	_, parserIsHeathy := pm.parsersStatusManager.GetParserStatus(source)

	// если в кэше ничего не было найдно, то выполняем запрос в конкретном парсере
	var parserForRequest interfaces.Parser
	//выбираем нужный парсер
	for _, parser := range pm.parsers {
		if parser.GetName() == source && parserIsHeathy == true {
			parserForRequest = parser
			break
		}
	}

	// делаем запрос выбранный сервис
	vacancyDetails, err := parserForRequest.SearchVacanciesDetailes(ctx, vacancyID)

	if err != nil {
		return models.SearchVacancyDetailesResult{}, err
	}

	// кэшируем результат в кэш для результатов поиска деталей вакансии по конкретному ID
	pm.cacheDetailsResult(vacancyID, vacancyDetails)

	return vacancyDetails, nil
}

// метод получения имени источника и ID вакансии из ввода
func (pm *ParsersManager) getCompositeIDFromInput(scanner *bufio.Scanner) (string, string, error) {
	fmt.Print("Введите ID вакансии: ")
	if !scanner.Scan() {
		return "", "", fmt.Errorf("❌ Проблема со сканированием ввода\n")
	}

	// переменная, куда сохранаяется ID искомой вакансии
	vacancyID := strings.TrimSpace(scanner.Text())
	if vacancyID == "" {
		//fmt.Println("❌ ID вакансии не может быть пустым")
		return "", "", fmt.Errorf("❌ ID вакансии не может быть пустым\n")
	}

	fmt.Print("Введите источник (HH.ru/SuperJob.ru): ")
	if !scanner.Scan() {
		return "", "", fmt.Errorf("❌ ввели неверное имя сервиса\n")
	}
	// переменная, куда кладём имя сервиса, в результатах поиска которого будем искать ID вакансии
	source := strings.TrimSpace(scanner.Text())

	return source, vacancyID, nil
}

// функция вывода в консоль данных о найденой вакансии
func printVacancyDetails(vacancy models.Vacancy) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("recovered from PANIC: [ ", rec, " ]")
		}
	}()

	fmt.Println("\n" + strings.Repeat("=", 50))

	fmt.Printf("🏢 %s\n", vacancy.Job)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("💼 Работодатель: %s\n", vacancy.Company)

	// проверяем поле salary на nil, чтобы не словить панику
	if vacancy.Salary == nil {
		fmt.Printf("💰 Зарплата: %s\n", "Salary is nil")
	} else {
		fmt.Printf("💰 Зарплата: %s\n", *vacancy.Salary)
	}

	fmt.Printf("📍 Местоположение: %s\n", vacancy.Location)
	//fmt.Printf("🕐 Опубликовано: %s\n", formatDate(vacancy.PublishedAt))
	fmt.Printf("🔗 Ссылка: %s\n", vacancy.URL)
	fmt.Printf("🆔 ID: %s\n", vacancy.ID)

	// Обрезаем описание для читаемости
	if len(vacancy.Description) > 1500 {
		vacancy.Description = vacancy.Description[:1500] + "..."
	}

	if vacancy.Description != "" {
		fmt.Println("\n📝 Описание:")
		fmt.Println(cleanHTML(vacancy.Description))
	}

	fmt.Println(strings.Repeat("=", 50))
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

// функция очистки HTML тегов из строки
func cleanHTML(text string) string {
	// Простая очистка HTML тегов
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<li>", "• ")

	// Удаляем HTML теги
	var result strings.Builder
	var inTag bool

	for _, ch := range text {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}

	return strings.TrimSpace(result.String())
}
