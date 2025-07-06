package saveText

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	mock_saveText "sh42ers/internal/http-server/handlers/url/saveText/mocks"

	//mapstorage "sh42ers/internal/storage/map"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Я проверяю только правильность методов.
//
// ВЕРНУТЬ К СТАРОМУ СПОСОБУ- с ерундой (это тупиковая ветвь, вся работа в save)
// И главная ерунда у меня- args{mapstorage.NewURLStorage(make(map[string]string))}
// Прямое обращение к другому участку кода,
// который я сам хочу заменить (на sqlite)
func TestSaveTextNew(t *testing.T) {
	// //
	// ctrl := gomock.NewController(t)
	// mockStorage := *mock_save.NewMockURLSaver(ctrl)
	// //

	// type args struct {
	// 	urlSaver URLSaver
	// }

	//// СПОТЫКАЕТСЯ mock на проверке, когда все в порядке.
	//// "Плохой" метод сюда не заходит внутрь NewText
	//// Нужно переделать по Тузовским образцам!!

	tests := []struct {
		name string
		//ts     args
		method string
		want   int
	}{
		// {
		// 	name: "all good",
		// 	//ts:     args{mapstorage.NewURLStorage(make(map[string]string))},
		// 	method: "POST",
		// 	want:   http.StatusCreated,
		// },
		{
			name: "bad method",
			//ts:     args{mapstorage.NewURLStorage(make(map[string]string))},
			method: "GET",
			want:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Начинаю использовать моки
			ctrl := gomock.NewController(t)
			//mockStorage := mock_save.NewMockURLSaver(ctrl)
			mockStorage := mock_saveText.NewMockURLtextSaver(ctrl)
			//

			req := httptest.NewRequest(tt.method, "/", nil) // bytes.NewBufferString("https://practicum.yandex.ru/"))
			req.Header.Set("Content-Type", "text/plain")

			rr := httptest.NewRecorder()

			//handler := http.HandlerFunc(NewText(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.ts.urlSaver))
			handler := http.HandlerFunc(NewText(slog.New(slog.NewJSONHandler(os.Stdout, nil)), mockStorage))
			handler.ServeHTTP(rr, req)

			// Пакет tesify
			require.Equal(t, tt.want, rr.Code)

			//// Пакет testing
			// // Работает!
			// if status := rr.Code; status != tt.want {
			// 	t.Errorf("Want status '%d', got '%d'", tt.want, rr.Code)
			// }
		})
	}
}
