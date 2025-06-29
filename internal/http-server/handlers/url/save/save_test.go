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

// // Так не работает!
// // Взять за основу Тузова, но по примеру gomockExample!!!
// func TestSaveMock(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Create a mock object for the UserRepository interface
// 	mockRepo := mock_save.NewMockURLSaver(ctrl)

// 	// Set the expected behavior:
// 	// when GetUserByID is called with "1", return a user.
// 	// Установим ожидаемое поведение (и передадим ожидаемые значения):
// 	// когда GetUserByID вызывается с параметром "1", возвращается имя пользователя.
// 	mockRepo.EXPECT().
// 	SaveURL("https://practicum.yandex.ru/","6ba7b811").
// 	Return()
// 		// GetUserByID("1").
// 		// Return(&models.User{ID: "1", Name: "Alex"}, nil)

// 	// serviceTest := service.NewUserService(mockRepo)
// 	// user, err := serviceTest.GetUser("1")

// 	// Я переделал в одну строчку. Мне так понятнее, а сути не меняет
// 	// Здесь вызов тестируемого метода GetUser
// 	// Если параметр "1", то все будет в порядке.
// 	user, err := service.NewUserService(mockRepo).GetUser("1")

// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	if user.ID != "1" {
// 		t.Errorf("Expected user ID to be '1', got %s", user.ID)
// 	}
// }

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

			handler := http.HandlerFunc(NewText(slog.New(slog.NewJSONHandler(os.Stdout, nil)), tt.ts.urlSaver))
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
