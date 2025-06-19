package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sh42ers/internal/config"
	handlers "sh42ers/internal/http-server/handlers"
	"sh42ers/internal/storage"
	"syscall"
	"time"

	mwLogger "sh42ers/internal/http-server/middleware/logger"
	"sh42ers/internal/lib/logger/sl"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Объявить переменные окружения из iter5 так:
// $env:SERVER_ADDRESS = "localhost:8089"
// $env:BASE_URL  = "http://localhost:9999"

// Если использую local.yaml , то перед запуском нужно установить переменную окружения CONFIG_PATH
//
// $env:CONFIG_PATH = "C:\__git\adv-url-shortener\config\local.yaml"
// $env:CONFIG_PATH = "C:\Mega\__git\adv-url-shortener\config\local.yaml"  (на ноуте)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// обрабатываем аргументы командной строки
	config.ParseFlags()

	// if err := Run(); err != nil {
	// 	panic(err)
	// }

	cfg := config.MustLoad()

	//
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env)) // к каждому сообщению будет добавляться поле с информацией о текущем окружении

	log.Info("initializing server", slog.String("address", cfg.Address)) // Помимо сообщения выведем параметр с адресом
	log.Debug("logger debug mode enabled")

	//
	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Logger)    // Логирование всех запросов
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	//переопределяем внутренний логгер
	router.Use(mwLogger.New(log))
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов

	// // adv #start#
	// router.Post("/", save.New(log, storage))
	// // adv #end#

	// Примитивное (based on map) хранилище
	storageInstance := storage.NewURLStorage(make(map[string]string))

	// // sqlite.New или "подключает" файл db , а если его нет то создает
	// storageInstance, err := sqlite.New("./storage.db")
	// if err != nil {
	// 	log.Error("failed to initialize storage", sl.Err(err))
	// }

	// routers
	//
	// В Go передача интерфейса параметром в функцию означает,
	// что функция может принимать на вход объект любого типа,
	// который реализует определенный интерфейс.
	//
	// PostHandler принимает параметром интерфейс URLSaver
	// с единственным методом SaveURL(URL, alias string) error
	// т.е. два строковых значения .
	// НО! Самое важное- то, что мы передадим параметром должно
	// реализовывать МЕТОДЫ интерфейса!

	router.Post("/", handlers.PostHandler(log, storageInstance))
	// GetHandler принимает ...
	router.Get("/{id}", handlers.GetHandler(log, storageInstance))

	// // примитивный запуск сервера
	//return http.ListenAndServe(config.FlagRunAddr, router)

	// adv #server#
	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

// // инициализации зависимостей сервера перед запуском
// func Run() error {
// 	router := chi.NewRouter()

// 	// storages

// 	// Примитивное (based on map) хранилище
// 	storageInstance := storage.NewURLStorage(make(map[string]string))

// 	// routers
// 	//
// 	// В Go передача интерфейса параметром в функцию означает,
// 	// что функция может принимать на вход объект любого типа,
// 	// который реализует определенный интерфейс.
// 	//
// 	// PostHandler принимает параметром интерфейс URLSaver
// 	// с единственным методом SaveURL(URL, alias string) error
// 	// т.е. два строковых значения .
// 	// НО! Самое важное- то, что мы передадим параметром должно
// 	// реализовывать МЕТОДЫ интерфейса!
// 	router.Post("/", handlers.PostHandler(storageInstance))
// 	// GetHandler принимает ...
// 	router.Get("/{id}", handlers.GetHandler(storageInstance))

// 	// server
// 	//fmt.Println("Running server on", flagRunAddr)
// 	return http.ListenAndServe(config.FlagRunAddr, router)
// }

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
