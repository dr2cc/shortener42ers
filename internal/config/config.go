package config

import (
	"flag"
	"os"
)

// переменная FlagRunAddr содержит адрес и порт для запуска сервера
var FlagRunAddr string

// переменная FlagURL отвечает за базовый адрес результирующего сокращённого URL
var FlagURL string

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменные
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagURL, "b", "http://localhost:8080", "host and port")
	// разбираем переданные серверу аргументы коммандной строки в зарегистрированные переменные
	flag.Parse()

	// Добавляем переменные окружения
	// $env:SERVER_ADDRESS = "localhost:8089"
	// $env:BASE_URL  = "http://localhost:9999"
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}
	if envURL := os.Getenv("BASE_URL"); envURL != "" {
		FlagURL = envURL
	}
}
