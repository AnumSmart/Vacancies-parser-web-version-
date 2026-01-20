package main

import (
	"context"
)

func main() {
	// Создаем корневой контекст
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

}
