package configs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"shared/config"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	API         APIConfig
	Cache       *CachesConfig
	Parsers     *ParsersConfig
	Manager     *ParserManagerConfig
	HealthChech *HealthCheckConfig
	Server      *config.ServerConfig
}

type APIConfig struct {
	ConcSearchTimeout time.Duration
	ServerPort        string
}

// загружаем конфиг-данные из .env
func LoadConfig() (*Config, error) {
	err := godotenv.Load("c:\\Son_Alex\\GO_projects\\go_v_1_23_web\\vacancy_parser\\search_service\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	concSearchTimeOut, err := strconv.Atoi(os.Getenv("CONC_SEARCH_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	cacheConfig, err := LoadYAMLConfig[CachesConfig](os.Getenv("CACHES_CONFIG_ADDRESS_STRING"), DefaultCacheConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersConfig, err := LoadYAMLConfig[ParsersConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersManagerConfig, err := LoadYAMLConfig[ParserManagerConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersManagerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	healthCheckConfig, err := LoadYAMLConfig[HealthCheckConfig](os.Getenv("HEALTH_CHECK_CONFIG_ADDRESS_STRING"), DefaultHealthCheckConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	serverConfig, err := LoadYAMLConfig[config.ServerConfig](os.Getenv("SERVER_CONFIG_ADDRESS_STRING"), config.DefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &Config{
		API: APIConfig{
			ConcSearchTimeout: time.Duration(concSearchTimeOut) * time.Second,
		},
		Cache:       cacheConfig,
		Parsers:     parsersConfig,
		Manager:     parsersManagerConfig,
		HealthChech: healthCheckConfig,
		Server:      serverConfig,
	}, nil
}

// универсальня функция загрузки конфига из .yml файла (используем дженерики, так как будут ещё парсеры)
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
