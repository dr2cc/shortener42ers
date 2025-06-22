package save

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	mapstorage "sh42ers/internal/storage/map"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveNew(t *testing.T) {
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
			ts:     args{mapstorage.NewURLStorage(make(map[string]string))},
			method: "POST",
			want:   http.StatusCreated,
		},
		{
			name:   "bad method",
			ts:     args{mapstorage.NewURLStorage(make(map[string]string))},
			method: "GET",
			want:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil) // bytes.NewBufferString("https://practicum.yandex.ru/"))
			req.Header.Set("Content-Type", "text/plain")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(New(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.ts.urlSaver))
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
