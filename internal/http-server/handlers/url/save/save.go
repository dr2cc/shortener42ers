package save

import (
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

// // TODO: move to config if needed
// const aliasLength = 6
// // Я и отправил в config, теперь это
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

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
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
			alias = random.NewRandomString(config.AliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		// возвращаемая Яндекс строка
		strResp := "http://" + r.Host + "/" + alias
		//

		//responseOK(w, r, alias)
		responseOK(w, r, strResp)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		//	Response: resp.OK(),
		Alias: alias,
	})
}
