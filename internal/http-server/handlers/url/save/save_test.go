package save

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	mock_save "sh42ers/internal/http-server/handlers/url/save/mocks"

	//mapstorage "sh42ers/internal/storage/map"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

// Тузовский
// func TestSaveHandler(t *testing.T) {
// 	cases := []struct {
// 		name      string
// 		alias     string
// 		url       string
// 		respError string
// 		mockError error
// 	}{
// 		{
// 			name:  "Success",
// 			alias: "test_alias",
// 			url:   "https://google.com",
// 		},
// 		{
// 			name:  "Empty alias",
// 			alias: "",
// 			url:   "https://google.com",
// 		},
// 		{
// 			name:      "Empty URL",
// 			url:       "",
// 			alias:     "some_alias",
// 			respError: "field URL is a required field",
// 		},
// 		{
// 			name:      "Invalid URL",
// 			url:       "some invalid URL",
// 			alias:     "some_alias",
// 			respError: "field URL is not a valid URL",
// 		},
// 		{
// 			name:      "SaveURL Error",
// 			alias:     "test_alias",
// 			url:       "https://google.com",
// 			respError: "failed to add url",
// 			mockError: errors.New("unexpected error"),
// 		},
// 	}

// 	for _, tc := range cases {
// 		tc := tc

// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			urlSaverMock := mocks.NewURLSaver(t)

// 			if tc.respError == "" || tc.mockError != nil {
// 				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
// 					Return(int64(1), tc.mockError).
// 					Once()
// 			}

// 			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

// 			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

// 			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
// 			// NoError проверяет, что функция не вернула ошибку.
// 			require.NoError(t, err)

// 			rr := httptest.NewRecorder()
// 			handler.ServeHTTP(rr, req)

// 			//Equal производит сравнение двух значений
// 			require.Equal(t, rr.Code, http.StatusOK)

// 			body := rr.Body.String()

// 			var resp save.Response
// 			// NoError проверяет, что функция не вернула ошибку.
// 			require.NoError(t, json.Unmarshal([]byte(body), &resp))
// 			//Equal производит сравнение двух значений
// 			require.Equal(t, tc.respError, resp.Error)

// 			// TODO: add more checks
// 		})
// 	}
// }

// Я проверяю только правильность методов.
// И главная ерунда у меня- args{mapstorage.NewURLStorage(make(map[string]string))}
// Прямое обращение к другому участку кода,
// который я сам хочу заменить (на sqlite)
func TestSaveNew(t *testing.T) {
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
		{
			name: "all good",
			//ts:     args{mapstorage.NewURLStorage(make(map[string]string))},
			method: "POST",
			want:   http.StatusCreated,
		},
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
			mockStorage := mock_save.NewMockURLSaver(ctrl)
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
