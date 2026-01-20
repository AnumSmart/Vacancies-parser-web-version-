package core

import (
	"fmt"
	"runtime"
	"search_service/configs"
)

// Dependencies содержит все общие зависимости
type AuthServiceDepenencies struct {
	Config *configs.Config
}

// InitDependencies инициализирует общие зависимости для auth_service
func InitDependencies() (*AuthServiceDepenencies, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	//--------------------------------------------- ДОРАБОТАТЬ!---------------------------------------------
	return nil, nil
}
