package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	maps "sh42ers/internal/storage"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	//Здесь общие для всех тестов данные
	shortURL := "6ba7b811"
	record := map[string]string{shortURL: "https://practicum.yandex.ru/"}

	tests := []struct {
		name       string
		method     string
		input      *maps.URLStorage
		want       string
		wantStatus int
	}{
		{
			name:   "all good",
			method: http.MethodGet,
			input: &maps.URLStorage{
				Data: record,
			},
			want:       "https://practicum.yandex.ru/",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "with bad method",
			method: http.MethodPost,
			input: &maps.URLStorage{
				Data: record,
			},
			want:       "Method not allowed",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "key in input does not match /6ba7b811",
			method: http.MethodGet,
			input: &maps.URLStorage{
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

			handler := http.HandlerFunc(GetHandler(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.input))
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

func TestPostHandler(t *testing.T) {
	type args struct {
		urlSaver URLSaver
	}

	tests := []struct {
		name   string
		ts     args
		method string
		want   int
	}{
		{
			name:   "all good",
			ts:     args{maps.NewURLStorage(make(map[string]string))},
			method: "POST",
			want:   http.StatusCreated,
		},
		{
			name:   "bad method",
			ts:     args{maps.NewURLStorage(make(map[string]string))},
			method: "GET",
			want:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil) // bytes.NewBufferString("https://practicum.yandex.ru/"))
			req.Header.Set("Content-Type", "text/plain")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(PostHandler(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.ts.urlSaver))
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
