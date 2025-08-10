// Package utils provides utility functions for the agent.
package utils

import (
	"bytes"
	"compress/gzip"
)

func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		closeErr := writer.Close()
		if closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
