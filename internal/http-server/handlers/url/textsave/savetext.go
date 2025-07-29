package savetext

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sh42ers/internal/config"
	"sh42ers/internal/lib/random"
)

// // До go generate нужно установить библиотеку
// // mockery имеет сложную установку, реализовал в POST json ендпойнте
//
//go:generate mockgen -source=saveText.go -destination=mocks/mock.go
type URLtextSaver interface {
	// // Метод SaveURL реализуется в обоих хранилищах- maps и sqlite
	SaveURL(urlToSave string, alias string) error
}

// //**////////////////////////////////////////

// type Shortener struct {
// 	//Random     random.Generator
// 	repository Repository
// 	//generator  generator.URLGenerator
// 	//config     *config.Config
// }

// type ShortURL struct {
// 	DeletedAt     time.Time `json:"deleted_at"` // is used to mark a record as deleted
// 	OriginalURL   string    `json:"url"`        // original URL that was shortened
// 	ID            string    `json:"id"`         // unique ID of the short URL.
// 	CreatedByID   string    `json:"created_by"` // ID of the user who created the short URL
// 	CorrelationID string    `json:"correlation_id"`
// }

// // Repository saves and retrieves data from storage.
// type Repository interface {
// 	Save(ctx context.Context, shortURL ShortURL) error
// 	GetByID(ctx context.Context, id string) (ShortURL, error)
// 	// GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error)
// 	// Close(_ context.Context) error
// 	// Check(ctx context.Context) error
// 	// SaveBatch(ctx context.Context, batch []models.ShortURL) error
// 	// DeleteUrls(ctx context.Context, urls []models.ShortURL) error
// 	// GetUsersAndUrlsCount(ctx context.Context) (int, int, error)
// }

// // Из нового проекта, отсюда вызывается Save, для записи в текстовый файл
// // Shorten shortens full url and returns filled struct ShortURL.
// func (service *Shortener) Shorten(ctx context.Context, url string, userID string) (models.ShortURL, error) {

// 	urlID, err := service.generator.GenerateIDFromString(url)
// 	if err != nil {
// 		return ShortURL{}, err
// 	}

// 	shortURL := ShortURL{
// 		OriginalURL: url,
// 		ID:          urlID,
// 		CreatedByID: userID,
// 	}

// 	fmt.Println("Save вызывается из Shorten!")
// 	err = service.repository.Save(ctx, shortURL)

// 	// // это потом, сейчас не провереяю уникальность
// 	// var notUniqueErr *storage.NotUniqueURLError
// 	// if errors.As(err, &notUniqueErr) {
// 	// 	return shortURL, NewShorteningError(shortURL, err)
// 	// }
// 	if err != nil {
// 		return ShortURL{}, err
// 	}

// 	return shortURL, nil
// }

// //**//////////////////////

func New(log *slog.Logger, urlSaver URLtextSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {

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

			// Объект urlSaver (переданный при создании хендлера из main)
			// используется именно тут!
			err = urlSaver.SaveURL(url, alias)
			//if urlSaver.SaveURL(url, alias) != nil {
			if err != nil {
				fmt.Println("failed to add url")
				return
			}

			// //** FileRepository **////////////////////////
			// // в начале буду записывать файл здесь
			// // вероятно надо сделать выбор в main, где выбор хранилища
			// err = filerepo..Save(ctx, shortURL)
			// if err != nil {
			// 	fmt.Println("failed to add url")
			// 	return
			// }
			// //** FileRepository **////////////////////////

			log.Info("url added", slog.String("alias", alias))

			// Устанавливаем статус ответа 201
			w.WriteHeader(http.StatusCreated)
			// fmt.Fprint(w, config.FlagURL+"/"+alias)
			if config.FlagURL == "none" {
				fmt.Fprint(w, "http://"+r.Host+"/"+alias)
			} else {
				fmt.Fprint(w, config.FlagURL+"/"+alias)
			}

		} else {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	}
}
