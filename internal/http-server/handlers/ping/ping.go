package ping

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	//_ "github.com/lib/pq"
)

func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// делаем запрос
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		// не забываем освободить ресурс
		defer cancel()

		// В принципе только для целей проверки наличия соединения достаточно db.Ping
		// но он тоже получает контест, только не явно.
		// Яндекс советует всюду использовать context
		// И уже все готово для работы с данными!
		err := db.PingContext(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			//fmt.Fprint(w, "Error connecting to the database:", err)
			return
		}
		//w.WriteHeader(http.StatusOK)
		//fmt.Fprint(w, dn, " - successfully connected to the database!")
		defer db.Close()
	}
}

// // iterXX? Более сложная, но best practice проверка "здоровья"
// // Вдруг пригодиться..
// func (a *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	// test connect to db
// 	if err := a.DB.PingContext(ctx); err != nil {
// 		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to DB: %v", err))
// 		return
// 	}

// 	// Executing a test query
// 	var result int
// 	if err := a.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
// 		respondWithError(w, http.StatusServiceUnavailable, fmt.Sprintf("Test request error: %v", err))
// 		return
// 	}

// 	respondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "available",
// 		"message": fmt.Sprintf("The test request returned: %d", result),
// 	})
// }

// // Helper (вспомогательные) functions for answers
// func respondWithError(w http.ResponseWriter, code int, message string) {
// 	respondWithJSON(w, code, map[string]string{"error": message})
// }

// func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	json.NewEncoder(w).Encode(payload)
// }
