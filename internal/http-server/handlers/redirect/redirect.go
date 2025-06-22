package redirect

import (
	"log/slog"
	"net/http"
	"strings"
)

// В Go передача интерфейса параметром в функцию означает,
// что функция может принимать на вход объект любого типа,
// который реализует определенный интерфейс.
type URLGetter interface {
	// Метод GetURL реализуется в обоих хранилищах- maps и sqlite
	// Так они оба реализуют интерфейс URLGetter
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGeter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			alias := strings.TrimPrefix(r.RequestURI, "/")

			url, err := urlGeter.GetURL(alias)
			if err != nil {
				w.Header().Set("Location", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		default:
			w.Header().Set("Location", "Method not allowed")
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
