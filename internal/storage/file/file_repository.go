package filerepo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

// ShortURL is main entity for system.
type ShortURL struct {
	CreatedByID string `json:"uuid"`
	ID          string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	//DeletedAt     time.Time `json:"deleted_at"`
	//CorrelationID string    `json:"correlation_id"`
}

// FileRepository is repository that uses files for storage.
type FileRepository struct {
	file   *os.File      // file that we will be writing to
	writer *bufio.Writer // buffered writer that will write to the file
	// // in furure
	//mutex  sync.RWMutex  // mutex that will be used to synchronize access to the file
}

// Constructor function (Factory function) NewFileRepository creates a new file repository.
func NewFileRepository(filePath string) (*FileRepository, error) {
	fmt.Println("конструктор NewFileRepository")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o777) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// Define a method ReadFileToMap for the FileRepository type with a pointer receiver - repo
// ReadFileToMap reads the file and returns a map of all the urls in the file.
func (repo *FileRepository) ReadFileToMap(existingURLs map[string]string) error {
	fmt.Println("метод (типа FileRepository) readFileToMap")
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	var entry ShortURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return err
		}
		existingURLs[entry.ID] = entry.OriginalURL
	}
	return nil
}

// Save saving URL & shortURL  to the file
func (repo *FileRepository) Save(URL string, shortURL string) error {
	// // пока нет задания на уникальность записей,
	// _, err := repo.GetByID(shortURL)
	// if err == nil {
	// 	//return NewNotUniqueURLError(shortURL, nil)
	// }

	//считаю строки
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	scanner := bufio.NewScanner(repo.file)
	count := 1

	for scanner.Scan() {
		count += 1
	}
	//

	jsonUrl := ShortURL{
		OriginalURL: URL,
		ID:          shortURL,
		CreatedByID: strconv.Itoa(count),
	}

	data, err := json.Marshal(jsonUrl)
	if err != nil {
		return err
	}

	// // in furure
	// repo.mutex.Lock()
	// defer repo.mutex.Unlock()

	if _, errWrite := repo.writer.Write(data); errWrite != nil {
		return errWrite
	}

	if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
		return errWriteByte
	}

	if errFlush := repo.writer.Flush(); errFlush != nil {
		return errFlush
	}

	return nil
}

// // GetByID gets url by id.
// // Reads the file line by line and returns url that matches given id.
// func (repo *FileRepository) GetByID(id string) (ShortURL, error) {
// 	// repo.mutex.RLock()
// 	// defer repo.mutex.RUnlock()

// 	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
// 		return ShortURL{}, err
// 	}

// 	var entry ShortURL

// 	scanner := bufio.NewScanner(repo.file)

// 	for scanner.Scan() {
// 		line := scanner.Bytes()
// 		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
// 			return ShortURL{}, err
// 		}

// 		if entry.ID == id {
// 			return entry, nil
// 		}
// 	}

// 	return ShortURL{}, errors.New("can't find full url by id")
// }

// // Close closes file.
// func (repo *FileRepository) Close(_ context.Context) error {
// 	fmt.Println("метод (типа FileRepository) Close")
// 	return repo.file.Close()
// }

// // Check checks if file is ok.
// func (repo *FileRepository) Check(_ context.Context) error {
// 	fmt.Println("метод (типа FileRepository) Check")
// 	_, err := repo.file.Stat()
// 	return err
// }

// // writeMapToFile writes the map to the file.
// func (repo *FileRepository) writeMapToFile(existingURLs map[string]models.ShortURL) error {
// 	fmt.Println("метод (типа FileRepository) writeMapToFile")
// 	if err := repo.file.Truncate(0); err != nil {
// 		return err
// 	}
// 	if _, err := repo.file.Seek(0, 0); err != nil {
// 		return err
// 	}

// 	for _, url := range existingURLs {

// 		data, err := json.Marshal(url)
// 		if err != nil {
// 			return err
// 		}

// 		if _, errWrite := repo.writer.Write(data); errWrite != nil {
// 			return errWrite
// 		}

// 		if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
// 			return errWriteByte
// 		}

// 	}
// 	if err := repo.writer.Flush(); err != nil {
// 		return err
// 	}
// 	return nil
// }
