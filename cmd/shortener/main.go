package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sh42ers/internal/config"
	"sh42ers/internal/http-server/handlers/redirect"
	"sh42ers/internal/http-server/handlers/url/save"
	savetext "sh42ers/internal/http-server/handlers/url/textsave"
	filerepo "sh42ers/internal/storage/file"
	mapstorage "sh42ers/internal/storage/map"
	"syscall"
	"time"

	"sh42ers/internal/http-server/middleware/compress"
	myLog "sh42ers/internal/http-server/middleware/logger"
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

//

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

	log.Info("init server", slog.String("address", cfg.Address)) // Помимо сообщения выведем параметр с адресом
	log.Debug("logger debug mode enabled")

	//
	router := chi.NewRouter()

	// Middleware встроенный в chi

	// Трассировка. Добавляется request_id в каждый запрос
	router.Use(middleware.RequestID)
	// Логирование всех запросов
	router.Use(middleware.Logger)
	// Если внутри произойдет паника, приложение не упадет.
	// Recoverer это compress.Gzipper, которое восстанавливается после паники,
	// регистрирует панику и выводит идентификатор запроса, если он указан.
	router.Use(middleware.Recoverer)

	// Меняю логгер на мой
	router.Use(myLog.New(log))
	// // Работа с gzip
	// // Можно включить, но этот middleware не фильтрует контент по типам и эндпойнтам!
	// // Применил его к конкретным эндпойнтам
	// router.Use(compress.Gzipper)
	// //
	// // а это встроенный в chi "компрессор" !
	// // Compress — это middleware, которое сжимает тело ответа с заданными типами контента
	// //  в формат данных на основе заголовка запроса Accept-Encoding. Используется заданный уровень сжатия.
	// //
	// // ПРИМЕЧАНИЕ: обязательно установите заголовок Content-Type в ответе,
	// // иначе промежуточное ПО не будет сжимать тело ответа.
	// // Например, в обработчике следует установить w.Header().Set("Content-Type", http.DetectContentType(yourBody))
	// // или задать его вручную.
	// //
	// // Передача уровня сжатия 5 — разумное значение.
	// router.Use(middleware.Compress(flate.BestSpeed))
	// //

	// Парсер URLов поступающих запросов.
	// Удалит суффикс из пути маршрутизации и продолжит маршрутизацию
	router.Use(middleware.URLFormat)

	// Примитивное (based on map) хранилище
	// С июля 2025 не думаю, что пригодиться, но если вдруг..
	// Оказалось не так! До unit3 сказали оставаться на map
	mapRepository := make(map[string]string)

	// iter9 Тут создадим мапу из файла
	repo, err := filerepo.NewFileRepository("pip.json") //("./cmd/shortener/pip.json")
	if err != nil {
		panic(err)
	}
	err = repo.ReadFileToMap(mapRepository)
	if err != nil {
		panic(err)
	}
	//

	storageInstance := mapstorage.NewURLStorage(mapRepository) //(make(map[string]string))

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
	// Хендлер с методом POST принимает параметром интерфейс URLSaver
	// с единственным методом SaveURL(URL, alias string) error
	// т.е. два строковых значения .
	// НО! Самое важное- то, что мы передадим параметром должно
	// реализовывать МЕТОДЫ интерфейса!

	// JSON POST эндпоинт
	// который будет принимать в теле запроса JSON-объект
	// Можно использовать Use, а можно With . Пока вижу отличия только в синтаксисе
	router.Route("/api/shorten", func(r chi.Router) {
		r.Use(compress.Gzipper)
		r.Post("/", save.New(log, storageInstance))
		//r.With(compress.Gzipper).Post("/", save.New(log, storageInstance))
	})
	// // вариант без middleware для gzip
	// router.Post("/api/shorten", save.New(log, storageInstance))

	// TEXT POST эндпойнт
	router.Route("/", func(r chi.Router) {
		r.With(compress.Gzipper).Post("/", savetext.New(log, storageInstance))
	})
	// // вариант без middleware для gzip
	// router.Post("/", savetext.New(log, storageInstance))

	// TEXT GET эндпойнт
	router.Route("/{id}", func(r chi.Router) {
		r.With(compress.Gzipper).Get("/", redirect.New(log, storageInstance))
	})
	// // вариант без middleware для gzip
	// router.Get("/{id}", redirect.New(log, storageInstance))

	// // Пример роутера для применения middleware к конкретному роуту в Go с использованием Chi
	// // Работает так:
	// // Для применения middleware к конкретному роуту или группе роутов мы используем router.Route().
	// // Внутри router.Route() мы используем r.Use() для применения middleware (в данном случае compress.Gzipper) к группе роутов, начинающихся с /public
	// //Для применения middleware к конкретному роуту, мы используем r.With(compress.Gzipper).Get("/data", ...) для роута /public/data.
	// //
	// router.Route("/public", func(r chi.Router) {
	// 	r.With(compress.Gzipper).Get("/data", func(w http.ResponseWriter, r *http.Request) {
	// 		w.Write([]byte("Public data"))
	// 	})
	// })

	// servers
	//
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
