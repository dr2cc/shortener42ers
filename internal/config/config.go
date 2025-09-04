package config

import (
	"flag"
	"os"
	"time"
)

// // Здесь все переменные окружения, по мере их добавления в проект.
// $env:SERVER_ADDRESS = "localhost:8089"
// $env:FILE_STORAGE_PATH  = "aliases.json"
// $env:BASE_URL  = "http://localhost:9999"
// $env:DATABASE_DSN="postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable"

//const AliasLength = 6
// // Ревьюер первого спринта сказал перенести сюда
// // а второго вернуть назад в логику хендлеров

// переменная FlagRunAddr содержит адрес и порт для запуска сервера
var FlagRunAddr string

// переменная FlagURL отвечает за базовый адрес результирующего сокращённого URL
var FlagURL string

// переменная FlagFILE отвечает за путь к файлу с адресами и псевдонимами (aliases)
var FlagFile string

// переменная FlagDsn отвечает за DATABASE_DSN
var FlagDsn string

func getEnvOrDefault(envKey, defaultValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return defaultValue
}

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// Регистрируем переменные, используемые как флаги.
	//
	// ИСПОЛЬЗУЕТСЯ как параметр Addr при запуске сервера:
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	// Example of using command line flag:
	// go run .\cmd\shortener\main.go -f kid
	flag.StringVar(&FlagFile, "f", "pip.json", "file storage path")
	// Тут передается то, что будет выдвать в ответ (до алиаса) текстовый POST хендлер
	// в теории должен быть связан с FlagRunAddr
	// и производиться разбор строки (можно любую ерунду передать)
	flag.StringVar(&FlagURL, "b", "http://localhost:8080", "host and port")

	// По условию uter10 DATABASE_DSN получаем или из переменной окружения или задаем поле флага -d
	// Если изменят, то добавить в параметр value значение DATABASE_DSN
	flag.StringVar(&FlagDsn, "d", "postgres://user:password@host:port/dbname?sslmode=disable", "db DSN")
	// разбираем переданные серверу аргументы коммандной строки в зарегистрированные переменные
	flag.Parse()

	// if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
	// 	FlagRunAddr = envRunAddr
	// }
	// if envURL := os.Getenv("BASE_URL"); envURL != "" {
	// 	FlagURL = envURL
	// }
	// if envFile := os.Getenv("FILE_STORAGE_PATH"); envFile != "" {
	// 	FlagFile = envFile
	// }

	FlagRunAddr = getEnvOrDefault("SERVER_ADDRESS", FlagRunAddr)
	FlagURL = getEnvOrDefault("BASE_URL", FlagURL)
	FlagFile = getEnvOrDefault("FILE_STORAGE_PATH", FlagFile)
	FlagDsn = getEnvOrDefault("DATABASE_DSN", FlagDsn)
}

// adv
// Структуры для анмаршаллинга
type Config struct {
	Env         string `yaml:"env" env-default:"development"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"0.0.0.0:8888"`
	// iter9
	FileRepo string `yaml:"file_repo" env-default:"pip.json"`
	// iter10
	DBDsn       string        `yaml:"db_dsn" env-default:"postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// Здесь использование конфигурации в local.yaml и переменная окружения CONFIG_PATH
func MustLoad() *Config {
	// ya #start# - заполняем нашу структуру сразу
	cfg := Config{
		Env:         "local",
		StoragePath: "./storage.db",
		HTTPServer: HTTPServer{
			Address:  FlagRunAddr, //"localhost:8080",
			FileRepo: FlagFile,
			DBDsn:    FlagDsn,
			//Timeout:     5,
			//IdleTimeout: 60,
		},
	}
	// ya #end#

	// // adv #start#
	// //Если использую local.yaml , то перед запуском нужно установить переменную окружения CONFIG_PATH
	// // Переделать на относительный путь!
	// $env:CONFIG_PATH = "C:\__git\*MyRepository*\config\local.yaml"       (на drkk)
	// $env:CONFIG_PATH = "C:\Mega\__git\*MyRepository*\config\local.yaml"  (на ноуте)

	// // Если будут проблемы с переменной окружения, то писать путь так (\ экранируется \\):
	//configPath := "C:\\__git\\*MyRepository*\\config\\local.yaml"

	// // Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	// configPath := os.Getenv("CONFIG_PATH")
	// if configPath == "" {
	// 	log.Fatal("CONFIG_PATH environment variable is not set")
	// }

	// // Проверяем существование конфиг-файла
	// if _, err := os.Stat(configPath); err != nil {
	// 	log.Fatalf("error opening config file: %s", err)
	// }

	// // Читаем конфиг-файл и заполняем нашу структуру

	// var cfg Config

	// err := cleanenv.ReadConfig(configPath, &cfg)
	// fmt.Println(cfg)
	// if err != nil {
	// 	log.Fatalf("error reading config file: %s", err)
	// }
	// // adv #end#

	return &cfg
}
