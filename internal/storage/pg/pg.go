package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// App предоставляет основное приложение
type App struct {
	DB *sql.DB
}

// Инициализация подключения к PostgreSQL
func InitDB(log *slog.Logger) (*sql.DB, error) {
	// Getting DSN from environment variables
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Error("DATABASE_DSN not specified in env")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connection error: %v", err)
	}

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяю подключение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error to ping: %v", err)
	}

	return db, nil
}

// Обработчик проверки "здоровья" БД
func (a *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// test connect to db
	if err := a.DB.PingContext(ctx); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to DB: %v", err))
		return
	}

	// Executing a test query
	var result int
	if err := a.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, fmt.Sprintf("Test request error: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "available",
		"message": fmt.Sprintf("The test request returned: %d", result),
	})
}

// Helper (вспомогательные) functions for answers
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
