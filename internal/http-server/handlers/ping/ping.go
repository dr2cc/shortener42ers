package ping

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5436
	user     = "postgres"
	password = "qwerty"
	dbname   = "postgres"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	// Cтрока подключения
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Открываем соединение
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка подключения: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Устанавливаем таймаут для Ping
	db.SetConnMaxLifetime(time.Second * 5)

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка ping: %v", err), http.StatusInternalServerError)
		return
	}

	// Проверяем доступность простым запросом
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка запроса: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "PostgreSQL доступен! Результат теста: %d", result)
}
