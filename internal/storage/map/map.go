package mapstorage

import (
	"errors"
	"fmt"
	filerepo "sh42ers/internal/storage/file"
)

// тип URLStorage
type URLStorage struct {
	Data map[string]string
	// Dependency Injection (DI). iter9
	FileRepo *filerepo.FileRepository
}

// конструктор объектов URLStorage
// Конструкторами (инициализаторами) называются функции
// принимающие в качестве аргументов параметры структуры и
// возвращающие новый экземпляр структуры.
// По соглашению они записываются как
// newНазваниеСтруктуры()
func NewURLStorage(d map[string]string, f *filerepo.FileRepository) URLStorage {
	return URLStorage{
		Data:     d,
		FileRepo: f,
	}
}

// метод SaveURL объектов URLStorage
// реализует интерфейсы URLSaver (описан в save.go) и URLtextSaver (описан в savetext.go)
func (s URLStorage) SaveURL(url string, id string) error {
	const op = "storage.map.SaveURL"

	s.Data[id] = url

	// iter9
	// Организую запись в файл
	err := s.FileRepo.Save(url, id)
	//err = repo.Save(url, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
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
