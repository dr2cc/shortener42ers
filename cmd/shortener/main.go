package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sh42ers/internal/config"
	"sh42ers/internal/http-server/handlers/redirect"
	"sh42ers/internal/http-server/handlers/url/save"
	savetext "sh42ers/internal/http-server/handlers/url/textsave"
	mapstorage "sh42ers/internal/storage/map"
	"strings"
	"syscall"
	"time"

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

func Gzipper(h http.Handler) http.Handler {
	//return func(w http.ResponseWriter, r *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// // по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// // который будем передавать следующей функции
		// ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			fmt.Println("Supports Gzip!")
			// // оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			// cw := NewCompressWriter(w)
			// // меняем оригинальный http.ResponseWriter на новый
			// ow = cw
			// // не забываем отправить клиенту все сжатые данные после завершения middleware
			// defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			fmt.Println("Sends Gzip")
			// // оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			// cr, err := NewCompressReader(r.Body)
			// if err != nil {
			// 	w.WriteHeader(http.StatusInternalServerError)
			// 	return
			// }
			// // меняем тело запроса на новое
			// r.Body = cr
			// defer cr.Close()
		}

		// // передаём управление хендлеру
		// h.ServeHTTP(ow, r)

		// Пока мой мидлварь только определяет кто отправляет/ принимает gzip
		// Но вернуть управление необходимо!
		h.ServeHTTP(w, r)
	})
}

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

	router.Use(middleware.RequestID) // трейсинг? (добавлю request_id в каждый запрос)
	router.Use(middleware.Logger)    // логирование всех запросов
	// Честный мидлварь! Все три эндпойнта работают!
	router.Use(Gzipper)
	//
	router.Use(middleware.Recoverer) // если внутри произойдет паника, приложение не упадет
	//меняю логгер на мой
	router.Use(myLog.New(log))
	router.Use(middleware.URLFormat) // парсер URLов поступающих запросов

	// Примитивное (based on map) хранилище
	// С июля 2025 не думаю, что пригодиться, но если вдруг..
	// Оказалось не так! До unit3 сказали оставаться на map
	storageInstance := mapstorage.NewURLStorage(make(map[string]string))

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

	// К iter7 - эндпоинт POST /api/shorten,
	// который будет принимать в теле запроса JSON-объект
	router.Post("/api/shorten", save.New(log, storageInstance))

	// Текстовый POST эндпойнт
	router.Post("/", savetext.New(log, storageInstance))
	// Хендлер с методом GET принимает ...
	router.Get("/{id}", redirect.New(log, storageInstance))

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
