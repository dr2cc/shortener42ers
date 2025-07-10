package savetext

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sh42ers/internal/config"
	"sh42ers/internal/lib/random"

	"strings"
)

// // До go generate нужно установить библиотеку
// // mockery имеет сложную установку, реализовал в POST json ендпойнте
//
//go:generate mockgen -source=saveText.go -destination=mocks/mock.go
type URLtextSaver interface {
	// // Метод SaveURL реализуется в обоих хранилищах- maps и sqlite
	// // Но в maps я не делал id (для лога), а в sqlite он уже есть
	// // История их похожести закончилась
	//SaveURL(URL, alias string) error
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLtextSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			if strings.Contains(contentType, "text/plain") {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				// Преобразуем тело запроса (тип []byte) в строку:
				url := string(body)

				// // Генерируем короткий идентификатор и создаем запись в нашем хранилище
				// //config.FlagURL соответствует "http://" + req.Host если не использовать аргументы
				alias := random.NewRandomString(config.AliasLength)

				//// ЗДЕСЬ СПОТЫКАЕТСЯ mock на проверке, когда все в порядке.
				//// "Плохой" метод сюда не заходит!
				//// Нужно переделать по Тузовским образцам!!

				// Объект urlSaver (переданный при создании хендлера из main)
				// используется именно тут!
				id, err := urlSaver.SaveURL(url, alias)
				//if urlSaver.SaveURL(url, alias) != nil {
				if err != nil {
					fmt.Println("failed to add url")
					return
				}

				log.Info("url added", slog.Int64("id", id))

				// Устанавливаем статус ответа 201
				w.WriteHeader(http.StatusCreated)
				//fmt.Fprint(w, config.FlagURL+"/"+alias)
				if config.FlagURL == "none" {
					fmt.Fprint(w, "http://"+r.Host+"/"+alias)
				} else {
					fmt.Fprint(w, config.FlagURL+"/"+alias)
				}

			} else {
				http.Error(w, "Incorrect Content-Type. Expected text/plain", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	}
}
