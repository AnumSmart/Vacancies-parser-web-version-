package config

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// универсальня функция загрузки конфига из .yml файла (используем дженерики)
// fn - функция конструктор конфига
func LoadYAMLConfig[T any](configPath string, fn func() *T) (*T, error) {
	// Вызываем переданную функцию-конструктор для создания экземпляра конфигурации.
	// На этом этапе в config будут значения по умолчанию, заданные в конструкторе.
	// Это важно, так как если файл конфигурации отсутствует или пуст,
	// у нас всё равно будет работоспособная конфигурация.
	config := fn()

	// Если configPath == "" (пустая строка), сразу возвращаются дефолтные значения.
	if configPath == "" {
		return config, nil
	}

	// Если файл по указанному пути не существует, возвращаются дефолтные значения БЕЗ ошибки.
	if _, err := os.Stat(configPath); errors.Is(err, fs.ErrNotExist) {
		return config, nil
	}

	// Если файл существует, но его не удалось прочитать или распарсить — возвращается ошибка.
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// пробуем анмаршалить конфиг из yml файла в структуру нужного типа
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return config, nil
}
