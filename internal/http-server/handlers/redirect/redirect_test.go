package redirect

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	mapstorage "sh42ers/internal/storage/map"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedirectNew(t *testing.T) {
	//Здесь общие для всех тестов данные
	shortURL := "6ba7b811"
	record := map[string]string{shortURL: "https://practicum.yandex.ru/"}

	tests := []struct {
		name       string
		method     string
		input      *mapstorage.URLStorage
		want       string
		wantStatus int
	}{
		{
			name:   "all good",
			method: http.MethodGet,
			input: &mapstorage.URLStorage{
				Data: record,
			},
			want:       "https://practicum.yandex.ru/",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "with bad method",
			method: http.MethodPost,
			input: &mapstorage.URLStorage{
				Data: record,
			},
			want:       "Method not allowed",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "key in input does not match /6ba7b811",
			method: http.MethodGet,
			input: &mapstorage.URLStorage{
				Data: map[string]string{"6ba7b81": "https://practicum.yandex.ru/"},
			},
			want:       "URL with such id doesn't exist",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/"+shortURL, nil) // bytes.NewBufferString("https://practicum.yandex.ru/")) //body)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(New(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.input))
			handler.ServeHTTP(rr, req)

			//// Пакет testify
			// // Если нужно, то так обрабатывают ошибку в testfy
			// require.NoError(t, err)
			require.Equal(t, tt.wantStatus, rr.Code)
			require.Equal(t, tt.want, strings.TrimSpace(rr.Header()["Location"][0]))

			//// Пакет testing
			// if gotStatus := rr.Code; gotStatus != tt.wantStatus {
			// 	t.Errorf("Want status '%d', got '%d'", tt.wantStatus, gotStatus)
			// }

			// if gotLocation := strings.TrimSpace(rr.Header()["Location"][0]); gotLocation != tt.want {
			// 	t.Errorf("Want location'%s', got '%s'", tt.want, gotLocation)
			// }
		})
	}
}
