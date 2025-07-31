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

// Объявить переменные окружения из iter5 так:
// $env:SERVER_ADDRESS = "localhost:8089"
// $env:BASE_URL  = "http://localhost:9999"

// Если использую local.yaml , то перед запуском нужно установить переменную окружения CONFIG_PATH
//
// $env:CONFIG_PATH = "C:\__git\adv-url-shortener\config\local.yaml"
// $env:CONFIG_PATH = "C:\Mega\__git\adv-url-shortener\config\local.yaml"  (на ноуте)

// const (
// 	envLocal = "local"
// 	envDev   = "dev"
// 	envProd  = "prod"
// )

//

func main() {
	// обрабатываем аргументы командной строки
	config.ParseFlags()

	// if err := Run(); err != nil {
	// 	panic(err)
	// }

	cfg := config.MustLoad()

	// // adv #server#
	// log.Info("starting server", slog.String("address", cfg.Address))

	log, router := handlers.NewRouter(cfg)

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
