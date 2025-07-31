package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

//** Секция ОТПРАВКИ gzip сжатых данных **//

// gzipWriter и его методы.
// Тип gzipWriter реализует интерфейс http.ResponseWriter
// позволяет сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type gzipWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

// Close закрывает gzip.Writer и отправляет все данные из буфера.
func (g *gzipWriter) Close() error {
	return g.gw.Close()
}

func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.gw.Write(p)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		gw: gzip.NewWriter(w),
	}
}

//** Секция ПОЛУЧЕНИЯ сжатых gzip данных **//

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type gzipReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gr.Close()
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.gr.Read(p)
}

func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		gr: zr,
	}, nil
}

func Gzipper(h http.Handler) http.Handler {
	//return func(w http.ResponseWriter, r *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// копирую http.ResponseWriter
		lw := w

		// проверяем, что клиент поддерживает получение данных в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		clientGzip := strings.Contains(acceptEncoding, "gzip")
		if clientGzip {
			//fmt.Println("Supports gzip!")
			// оборачиваем http.ResponseWriter в gzipWriter
			ngw := NewGzipWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			lw = ngw
			// закрываю gzipWriter
			defer ngw.Close()
		}

		// проверяем, что сервер получает gzip сжатые данные
		contentEncoding := r.Header.Get("Content-Encoding")
		receivedGzip := strings.Contains(contentEncoding, "gzip")
		if receivedGzip {
			//fmt.Println("Received gzip!")
			// оборачиваем тело запроса в gzipReader
			ngr, err := NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = ngr
			defer ngr.Close()
		}

		// возвращаем управление обработчику
		h.ServeHTTP(lw, r)
	})
}
