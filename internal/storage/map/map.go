package mapstorage

import (
	"errors"
	filerepo "sh42ers/internal/storage/file"
)

// тип URLStorage .
// Покажи нам методы VSCode:
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

// Перед этим NewURLStorage записывал как ниже,
// новая запись более соответствует "канонам", а значит надежнее при развитии
// func NewURLStorage() URLStorage {
// 	return URLStorage{
// 		Data: make(map[string]string),
// 	}
// }

// iter9
// файл создал
func createFilerepo() *filerepo.FileRepository {
	repo, err := filerepo.NewFileRepository("pip.json") //("./cmd/shortener/pip.json")
	if err != nil {
		panic(err)
	}
	return repo
}

// метод SaveURL объектов URLStorage
// реализует интерфейс URLSaver (описан в handlers.go)
func (s URLStorage) SaveURL(url string, id string) error {
	s.Data[id] = url

	repo := createFilerepo()
	// Организую запись в него
	// Видимо логично сделать при записе в map !
	err := repo.Save(url, id)
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
