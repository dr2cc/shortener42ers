package mapstorage

import (
	"errors"
	"sh42ers/internal/config"
	filerepo "sh42ers/internal/storage/file"
)

// тип URLStorage
type URLStorage struct {
	Data map[string]string
}

// конструктор объектов URLStorage
// Конструкторами (инициализаторами) называются функции
// принимающие в качестве аргументов параметры структуры и
// возвращающие новый экземпляр структуры.
// По соглашению они записываются как
// newНазваниеСтруктуры()
func NewURLStorage(d map[string]string) URLStorage {
	return URLStorage{
		Data: d,
	}
}

// метод SaveURL объектов URLStorage
// реализует интерфейс URLSaver (описан в handlers.go)
func (s URLStorage) SaveURL(url string, id string) error {
	s.Data[id] = url

	// iter9 создаю файл или проверяю, что он существует
	// Вроде не совсем правильно.
	// После запуска сервера файл должен существовать
	//
	// Но остается вопрос- как сюда получить экземпляр FileRepository,
	// как не воспользовавшись NewFileRepository??
	cfg := config.MustLoad()
	repo, err := filerepo.NewFileRepository(cfg.FileRepo) //("pip.json") //("./cmd/shortener/pip.json")
	if err != nil {
		panic(err)
	}
	// Организую запись в него
	err = repo.Save(url, id)
	if err != nil {
		panic(err)
	}
	//

	return nil
}

// метод GetURL объектов URLStorage
// реализует интерфейс URLGeter (для Ян, описан в handlers.go)
func (s URLStorage) GetURL(id string) (string, error) {
	e, exists := s.Data[id]
	if !exists {
		return id, errors.New("URL with such id doesn't exist")
	}
	return e, nil
}
