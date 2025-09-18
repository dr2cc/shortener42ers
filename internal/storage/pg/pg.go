package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"sh42ers/internal/config"
	"time"

	_ "github.com/lib/pq"
)

// // iterXX? App предоставляет основное приложение
// // Более сложная, но best practice
// // Вдруг пригодиться..
// type App struct {
// 	DB *sql.DB
// }

type Storage struct {
	DB *sql.DB
}

// Инициализация подключения к PostgreSQL
func InitDB(log *slog.Logger) (*Storage, error) {

	cfg := config.MustLoad()
	// Getting DSN from environment variables
	//dsn := os.Getenv("DATABASE_DSN")
	dsn := cfg.DSN
	if dsn == "" {
		log.Error("DATABASE_DSN not specified in env")
		os.Exit(1)
	}

	// 1. Подключение к базе
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		// log.Error("DB connection error")
		// return false
		return nil, fmt.Errorf("connection error: %v", err)
	}

	// // Не забыть про defer!!
	//defer db.Close()

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяю подключение с таймаутом ответа
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// освобождаем ресурс
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error to ping: %v", err)
	}

	return &Storage{DB: db}, nil
}

func New(log *slog.Logger, db *sql.DB) error {
	const op = "storage.pg.New" // Имя текущей функции для логов и ошибок

	// 2. Создаем таблицу, если ее еще нет
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS aliases(
        alias VARCHAR NOT NULL UNIQUE,
        url TEXT NOT NULL);
	`)

	if err != nil {
		panic(err)
	}

	// Отправляем комманду (CREATE TABLE в данном случае)
	// Exec выполняет подготовленный оператор (stmt) с заданными аргументами
	// и возвращает [Result], суммирующий эффект оператора.
	// В данной ситуации этот "эффект" не используется
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}

	return nil
}
