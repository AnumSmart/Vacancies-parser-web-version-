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
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	// обработка возможной паники
	defer recoverWithDetails()

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

	// Остановка сервера
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	fmt.Println("👋 Сервер авторизации остановлен")

	// Остановка сервисов
	server.Handler.ShutDown(ctx)

}

func recoverWithDetails() {
	if r := recover(); r != nil {
		fmt.Printf("❌ ПАНИКА: %v\n", r)

		// Пропускаем первые 2 фрейма (recover и текущую defer функцию)
		pc := make([]uintptr, 10)
		n := runtime.Callers(3, pc)
		frames := runtime.CallersFrames(pc[:n])

		fmt.Println("📍 Стек вызовов:")
		i := 0
		for {
			frame, more := frames.Next()

			// Пропускаем runtime фреймы
			if !strings.Contains(frame.File, "runtime/") {
				fmt.Printf("  %d. %s\n", i, frame.Function)
				fmt.Printf("     %s:%d\n", frame.File, frame.Line)
				i++
			}

			if !more {
				break
			}
		}
	}
}
