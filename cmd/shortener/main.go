package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sh42ers/internal/config"
	"sh42ers/internal/http-server/handlers"
	"syscall"
	"time"

	"sh42ers/internal/lib/logger/sl"
)

// //Объявить переменные окружения:
// // PS
// $env:SERVER_ADDRESS="localhost:8089"
// $env:DATABASE_DSN="postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable"
// // bash
// export SERVER_ADDRESS="localhost:8089"
// export DATABASE_DSN="postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable"

func main() {
	// обрабатываем аргументы командной строки
	config.ParseFlags()

	// if err := Run(); err != nil {
	// 	panic(err)
	// }

	cfg := config.MustLoad()

	// // adv #server#
	// log.Info("starting server", slog.String("address", cfg.Address))

	// Создаю маршрутизатор.
	// В нем будет:
	// - логика подключения эндпойнтов
	// - логика подключения к хралилищам
	// - логика хранения и получения данных
	log, router := handlers.NewRouter(cfg)

	// Логика Graceful Shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Server startup parameters:
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
		//ReadTimeout:  cfg.HTTPServer.Timeout,
		//WriteTimeout: cfg.HTTPServer.Timeout,
		//IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Логика web-сервера
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")

}
