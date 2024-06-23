package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func Compress(data []byte) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	writer, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func Decompress(reader io.Reader) (*bytes.Buffer, error) {
	r, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer CloseWithLogging(r)

	var result bytes.Buffer
	_, err = result.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return &result, nil
}
