// описание общего конфига для сервиса авторизации
package configs

import (
	"fmt"
	"os"
	"shared/config"

	"github.com/joho/godotenv"
)

type AuthServiceConfig struct {
	ServerConf     *config.ServerConfig
	PostgresDBConf *config.PostgresDBConfig
}

// загружаем конфиг-данные из .env
func LoadConfig() (*AuthServiceConfig, error) {
	err := godotenv.Load("c:\\Son_Alex\\GO_projects\\go_v_1_20_web\\vacancy_parser\\auth_service\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .yml файла для serverConfig
	serverConfig, err := config.LoadYAMLConfig[config.ServerConfig](os.Getenv("SERVER_CONFIG_ADDRESS_STRING"), config.UseDefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .yml файла для postgresDBConfig
	postgresDBConfig, err := config.LoadYAMLConfig[config.PostgresDBConfig](os.Getenv("POSTGRES_CONFIG_ADDRESS_STRING"), config.UseDefautPostgresDBConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &AuthServiceConfig{
		ServerConf:     serverConfig,
		PostgresDBConf: postgresDBConfig,
	}, nil
}
