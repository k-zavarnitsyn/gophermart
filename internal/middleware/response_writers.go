package middleware

import (
	"bytes"
	"io"
	"net/http"
)

var _ http.ResponseWriter = (*loggingResponseWriter)(nil)

type CustomWriter struct {
	http.ResponseWriter

	Writer io.Writer
}

type responseData struct {
	data   *bytes.Buffer
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter

	responseData responseData
}

func newLoggingResponseWriter(writer http.ResponseWriter, withDataLog bool) *loggingResponseWriter {
	w := &loggingResponseWriter{
		ResponseWriter: writer,
		responseData: responseData{
			status: http.StatusOK,
		},
	}
	if withDataLog {
		w.responseData.data = &bytes.Buffer{}
	}

	return w
}

func (w *CustomWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if r.responseData.data != nil {
		r.responseData.data.Write(b)
	}

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
