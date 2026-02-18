package main

import (
	authserver "auth_service/internal/auth_server"
	"auth_service/internal/core"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Создаем корневой контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем общие зависимости
	deps, err := core.InitDependencies(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Создаем HTTP-сервер
	server, err := authserver.NewAuthServer(ctx, deps.AuthConfig.ServerConf, deps.AuthHandler)
	if err != nil {
		panic("Failed to create server!")
	}

	// создаём канал, который бдут реагировать на системные сигналы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		fmt.Printf("🚀 HTTP сервер авторизации запускается на %s\n", deps.AuthConfig.ServerConf.Addr())
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера авторизации...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер (ждем текущие запросы)
	fmt.Println("Останавливаем HTTP auth сервер...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Закрываем зависимости при выходе
	err = deps.Close()
	if err != nil {
		log.Printf("Error during resourses closing: %v", err)
	}

	fmt.Println("👋 Сервер авторизации остановлен")

}
