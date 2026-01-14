package converters

import (
	"fmt"
	"strings"
	"time"
)

// впомогательная функця получения имени источника
func getSourceName(source string) string {
	sourceMap := map[string]string{
		"hh":       "hh.ru",
		"superjob": "SuperJob",
		"habr":     "Habr Career",
		"zarplata": "Зарплата.ру",
	}

	if name, ok := sourceMap[source]; ok {
		return name
	}
	return source
}

// вспомогательная функция получения иконки
func getSourceIcon(source string) string {
	iconMap := map[string]string{
		"hh":       "https://hh.ru/favicon.ico",
		"superjob": "https://www.superjob.ru/favicon.ico",
		"habr":     "https://career.habr.com/favicon.ico",
	}

	if icon, ok := iconMap[source]; ok {
		return icon
	}
	return ""
}

// вспомогательная функция форматирования поля зарплаты
func formatSalary(salary *string, currency string) string {
	if salary == nil || *salary == "" {
		return "не указана"
	}
	return fmt.Sprintf("%s %s", *salary, getCurrencySymbol(currency))
}

// вспомогательная функция получения символа валюты
func getCurrencySymbol(currency string) string {
	switch strings.ToUpper(currency) {
	case "RUB", "RUR":
		return "₽"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	default:
		return currency
	}
}

// вспомогательная фукнция получения опыта
func formatExperience(exp string) string {
	// Маппинг кодов опыта в читаемый вид
	experienceMap := map[string]string{
		"noExperience": "Нет опыта",
		"between1And3": "1-3 года",
		"between3And6": "3-6 лет",
		"moreThan6":    "Более 6 лет",
	}

	if formatted, ok := experienceMap[exp]; ok {
		return formatted
	}
	return exp // возвращаем оригинал, если маппинга нет
}

// вспомогательная фукнция получения графика работы
func formatSchedule(schedule string) string {
	scheduleMap := map[string]string{
		"fullDay":     "Полный день",
		"shift":       "Сменный график",
		"flexible":    "Гибкий график",
		"remote":      "Удаленная работа",
		"flyInFlyOut": "Вахтовый метод",
	}

	if formatted, ok := scheduleMap[schedule]; ok {
		return formatted
	}
	return schedule
}

// вспомогательная фукнция получения даты публикации
func formatPublishedAt(publishedAt time.Time) string {
	now := time.Now()
	diff := now.Sub(publishedAt)

	switch {
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 0 {
			return "только что"
		}
		return fmt.Sprintf("%d %s назад", minutes, pluralize(minutes, "минуту", "минуты", "минут"))

	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d %s назад", hours, pluralize(hours, "час", "часа", "часов"))

	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d %s назад", days, pluralize(days, "день", "дня", "дней"))

	default:
		// Если больше месяца, показываем дату
		return publishedAt.Format("02.01.2006")
	}
}

// вспомогательная функция склонения в предложении
func pluralize(n int, singular, few, many string) string {
	n = n % 100
	if n >= 11 && n <= 19 {
		return many
	}

	switch n % 10 {
	case 1:
		return singular
	case 2, 3, 4:
		return few
	default:
		return many
	}
}

// Вспомогательная функция для форматирования времени
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
