package saveText

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sh42ers/internal/config"
	"sh42ers/internal/lib/random"
	"strings"
)

// // Не забыть, что до go generate нужно установить библиотеку
// // mockery имеет сложную установку, видимо этой библиотекой нужно пользоваться
// // уже на другом уровне
// // Еще момент! go run с описанием версии относится к go 1.17, не факт,что так работает в 1.23
// // Установку go install они не советуют (сайт https://vektra.github.io/mockery/latest/)
// // go install github.com/vektra/mockery/v2@v2.20.0
//go::generate go run github.com/vektra/mockery/v2@v2.20.0 --name=URLSaver

//go:generate mockgen -source=saveText.go -destination=mocks/mock.go
type URLtextSaver interface {
	// Метод SaveURL реализуется в обоих хранилищах- maps и sqlite
	SaveURL(URL, alias string) error
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
				if urlSaver.SaveURL(url, alias) != nil {
					fmt.Println("failed to add url")
					return
				}

				// id, err := urlSaver.SaveURL(url, alias)

				// if err != nil {
				// 	fmt.Println("failed to add url")
				// 	return
				// }

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
