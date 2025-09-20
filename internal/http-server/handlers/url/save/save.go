package save

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	resp "sh42ers/internal/lib/api/response"
	"sh42ers/internal/lib/logger/sl"
	"sh42ers/internal/lib/random"
	"sh42ers/internal/storage"
)

type Response struct {
	// // в iter7 у Яндекса одна строка
	// // тест сломается?
	// resp.Response
	Alias string `json:"result,omitempty"` //`json:"alias,omitempty"`
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

// // mockery так не работает. Очень быстро развивается! Теперь все в конфигурационном файле
//
// // docker run -v "$PWD":/src -w /src vektra/mockery:3
// // On PS: все, бросил, очень запутанная (06.06.2025)
// // docker run -v ${PWD}:/src -w /src vektra/mockery:3
// // docker run -v ${PWD}:/src -w /src vektra/mockery --all

// // вызов другой библиотеки генерации моков
//go::generate mockgen -source=save.go -destination=mocks/URLSaver.go

// Этот интерфейс описан в функциях
// storage.map.SaveURL
// storage.sqlite.SaveURL
// Нужно сделать такую для pg (по типу sqlite.SaveURL)
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) error
}

// // ПОЛУЧЕНИЕ сжатых gzip данных
// // перенес в middleware
// func handlingBody(r *http.Request) ([]byte, error) {
// 	contentEncoding := r.Header.Get("Content-Encoding")
// 	// проверяем через strings.Contains
// 	if !strings.Contains(contentEncoding, "gzip") {
// 		return io.ReadAll(r.Body)
// 	}
// 	gz, err := gzip.NewReader(r.Body)
// 	if err != nil {
// 		return []byte{}, err //обернуть ошибку?
// 	}
// 	defer gz.Close()
// 	return io.ReadAll(gz)
// }

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// // Добавил поддержку gzip, теперь r.Body нужно обрабатывать
		// body, fail := handlingBody(r)

		body, fail := io.ReadAll(r.Body)

		if fail != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		err := json.Unmarshal(body, &req)
		//err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// отдельно если тело запроса пустое
			log.Error("request body is empty")
			render.JSON(w, r, resp.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString() //(config.AliasLength)
		}

		// Здесь записываем в "хранилище"
		errURL := urlSaver.SaveURL(req.URL, alias)

		if errors.Is(errURL, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		// // Ищу "последнюю" точку где может быть записан w.WriteHeader
		// w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusCreated)

		jsonResp := Response{
			Alias: "http://" + r.Host + "/" + alias,
		}

		enc := json.NewEncoder(w)

		// Важен порядок!
		// После того как вызван w.WriteHeader(http.StatusCreated),
		// он уже не может записать соответствующий заголовок.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		// Видимо Encode совершает w.WriteHeader ..
		if err := enc.Encode(jsonResp); err != nil {
			log.Error("failed to add url", sl.Err(err))
			return
		}

		log.Info("url added", slog.String("alias", alias))

	}
}
