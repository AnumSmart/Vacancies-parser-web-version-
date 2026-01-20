// описание общего конфига для сервиса авторизации
package configs

import (
	"fmt"
	"shared/config"

	"github.com/joho/godotenv"
)

type Config struct {
	Server *config.ServerConfig
}

// загружаем конфиг-данные из .env
func LoadConfig() (*Config, error) {
	err := godotenv.Load("c:\\Users\\aliaksei.makarevich\\go\\go_v_1_20_web\\Job_Parser\\auth_service\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}
	//--------------------------------------------- ДОРАБОТАТЬ!---------------------------------------------
	return nil, nil
}
