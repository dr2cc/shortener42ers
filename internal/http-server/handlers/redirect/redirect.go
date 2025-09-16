package redirect

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// В Go передача интерфейса параметром в функцию означает,
// что функция может принимать на вход объект любого типа,
// который реализует определенный интерфейс.

// go::generate mockgen -source=redirect.go -destination=mocks/mock.go
type URLGetter interface {
	// Метод GetURL реализуется в обоих хранилищах- maps и sqlite
	// Так они оба реализуют интерфейс URLGetter
	GetURL(alias string) (string, error)
}

func NewDb(log *slog.Logger, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//switch r.Method
		if r.Method == http.MethodGet {
			//case http.MethodGet:
			alias := strings.TrimPrefix(r.RequestURI, "/")
			log.Info("endpoint TEXT GET", slog.String("alias", alias))

			// url, err := urlGeter.GetURL(alias)

			stmt, err := db.Prepare("SELECT url FROM aliases WHERE alias = $1")
			if err != nil {
				w.Header().Set("Location", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var resURL string

			err = stmt.QueryRow(alias).Scan(&resURL)
			if errors.Is(err, sql.ErrNoRows) {
				log.Error(err.Error())
				return //"", storage.ErrURLNotFound
			}

			w.Header().Set("Location", resURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			// w.Header().Set("Location", "Method not allowed")
			// w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func New(log *slog.Logger, urlGeter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//switch r.Method
		if r.Method == http.MethodGet {
			//case http.MethodGet:
			alias := strings.TrimPrefix(r.RequestURI, "/")

			url, err := urlGeter.GetURL(alias)
			if err != nil {
				w.Header().Set("Location", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			// w.Header().Set("Location", "Method not allowed")
			// w.WriteHeader(http.StatusBadRequest)
		}
	}
}
