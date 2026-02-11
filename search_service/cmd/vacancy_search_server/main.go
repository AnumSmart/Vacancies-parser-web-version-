package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"search_service/internal/core"
	"search_service/internal/search_server"
	"syscall"
	"time"
)

func main() {
	// обработка возможной паники
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Поймали панику:", r)
		}
	}()

	// Создаем корневой контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем общие зависимости
	deps, err := core.InitDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Создаем HTTP-сервер
	server, err := search_server.NewSearchServer(ctx, deps.Config.ServerConf, deps.SearchHandler)
	if err != nil {
		panic("Failed to create server!")
	}

	// создаём канал, который бдут реагировать на системные сигналы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		fmt.Printf("🚀 HTTP сервер поиска вакансий запускается на %s\n", deps.Config.ServerConf.Addr())
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера поиска вакансий...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Остановка сервера
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	fmt.Println("👋 Сервер поиска вакансий остановлен")

	// Остановка сервисов
	server.Handler.ShutDown(ctx)

}
