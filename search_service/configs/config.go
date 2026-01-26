// описание общего конфига для сервиса поиска
package configs

import (
	"fmt"
	"os"
	"shared/config"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type SearchServiceConfig struct {
	API         APIConfig
	Cache       *CachesConfig
	Parsers     *ParsersConfig
	Manager     *ParserManagerConfig
	HealthChech *HealthCheckConfig
	ServerConf  *config.ServerConfig
}

type APIConfig struct {
	ConcSearchTimeout time.Duration
	ServerPort        string
}

// загружаем конфиг-данные из .env
func LoadConfig() (*SearchServiceConfig, error) {
	err := godotenv.Load("c:\\Users\\aliaksei.makarevich\\go\\go_v_1_20_web\\Job_Parser\\search_service\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	concSearchTimeOut, err := strconv.Atoi(os.Getenv("CONC_SEARCH_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	cacheConfig, err := config.LoadYAMLConfig[CachesConfig](os.Getenv("CACHES_CONFIG_ADDRESS_STRING"), DefaultCacheConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersConfig, err := config.LoadYAMLConfig[ParsersConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersManagerConfig, err := config.LoadYAMLConfig[ParserManagerConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersManagerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	healthCheckConfig, err := config.LoadYAMLConfig[HealthCheckConfig](os.Getenv("HEALTH_CHECK_CONFIG_ADDRESS_STRING"), DefaultHealthCheckConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	serverConfig, err := config.LoadYAMLConfig[config.ServerConfig](os.Getenv("SERVER_CONFIG_ADDRESS_STRING"), config.UseDefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &SearchServiceConfig{
		API: APIConfig{
			ConcSearchTimeout: time.Duration(concSearchTimeOut) * time.Second,
		},
		Cache:       cacheConfig,
		Parsers:     parsersConfig,
		Manager:     parsersManagerConfig,
		HealthChech: healthCheckConfig,
		ServerConf:  serverConfig,
	}, nil
}
