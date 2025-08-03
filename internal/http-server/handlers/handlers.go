package handlers

import (
	"log/slog"
	"os"
	"sh42ers/internal/config"
	"sh42ers/internal/http-server/handlers/redirect"
	"sh42ers/internal/http-server/handlers/url/save"
	"sh42ers/internal/http-server/middleware/compress"

	savetext "sh42ers/internal/http-server/handlers/url/textsave"
	myLog "sh42ers/internal/http-server/middleware/logger"
	filerepo "sh42ers/internal/storage/file"
	mapstorage "sh42ers/internal/storage/map"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func NewRouter(cfg *config.Config) (*slog.Logger, *chi.Mux) {

	log := setupLogger(cfg.Env)
	//log = log.With(slog.String("env", cfg.Env)) // к каждому сообщению будет добавляться поле с информацией о текущем окружении

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
	repo, err := filerepo.NewFileRepository(cfg.FileRepo) //("pip.json") //("./cmd/shortener/pip.json")
	if err != nil {
		panic(err)
	}
	err = repo.ReadFileToMap(mapRepository)
	if err != nil {
		panic(err)
	}
	//

	storageInstance := mapstorage.NewURLStorage(mapRepository, repo) //(make(map[string]string))

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

	return log, router
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
