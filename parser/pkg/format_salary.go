package pkg

import "fmt"

func FormatSalary(from, to int, currency string) string {
	if from > 0 && to > 0 {
		return formatNumber(from) + " - " + formatNumber(to) + " " + currency
	} else if from > 0 {
		return "от " + formatNumber(from) + " " + currency
	} else {
		return "до " + formatNumber(to) + " " + currency
	}
}

// эта функция разделяет пробелами тысячи
func formatNumber(num int) string {
	if num >= 1000 {
		// Рекурсивно обрабатываем тысячи и добавляем пробел
		return formatNumber(num/1000) + " " + fmt.Sprintf("%03d", num%1000)
	}
	// Базовый случай - возвращаем число как строку
	return fmt.Sprintf("%d", num)
}
