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

	"sh42ers/internal/config"
	resp "sh42ers/internal/lib/api/response"
	"sh42ers/internal/lib/logger/sl"
	"sh42ers/internal/lib/random"
	"sh42ers/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	// // в iter7 у Яндекса одна строка
	// // тест сломается!
	// resp.Response
	Alias string `json:"result,omitempty"` //`json:"alias,omitempty"`
}

// const aliasLength = 6
// // Я отправил в config, теперь это
// // config.AliasLength

// // вызов другой библиотеки генерации моков
//go::generate mockgen -source=save.go -destination=mocks/URLSaver.go

// //не работает. По моему очень капризная или вообще не для windows .
// // В других проектах буду использовать gomock
//
// // docker run -v "$PWD":/src -w /src vektra/mockery:3
// // On PS: все, бросил, очень запутанная (06.06.2025)
// // docker run -v ${PWD}:/src -w /src vektra/mockery:3
// // docker run -v ${PWD}:/src -w /src vektra/mockery --all

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func renderJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		//ян// Нельзя пользоваться render , пробую json
		body, fail := io.ReadAll(r.Body)
		if fail != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		err := json.Unmarshal(body, &req)
		//*/
		//err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// отдельно если тело запроса пустое
			log.Error("request body is empty")
			//ян// все заменить на json
			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			//ян// все заменить на json
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			//ян// все заменить на json
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(config.AliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			//ян// все заменить на json
			render.JSON(w, r, resp.Error("url already exists"))

			return
		}
		// if err != nil {
		// 	log.Error("failed to add url", sl.Err(err))

		// 	//ян// все заменить на json
		// 	render.JSON(w, r, resp.Error("failed to add url"))

		// 	return
		// }

		log.Info("url added", slog.Int64("id", id))

		//11.07 Еще не превратил в json !!
		jsonResp := Response{
			Alias: "http://" + r.Host + "/" + alias,
		}
		// // Устанавливаем статус ответа 201
		// // После этой установки тип ответа становится text!
		// // Нам этого не нужно!
		// w.WriteHeader(http.StatusCreated)
		// // Эта конструкция ситуацию не исправляяет. Видимо статус нужно задавать в другом месте
		// w.Header().Set("Content-Type", "application/json")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(jsonResp); err != nil {
			//здесь будет мой логгер
			//logger.Log.Debug("error encoding response", zap.Error(err))
			log.Error("failed to add url", sl.Err(err))
			return
		}

	}
}
