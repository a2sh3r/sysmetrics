// Package utils provides utility functions for the agent.
package utils

import (
	"bytes"
	"compress/gzip"
	"net"
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

// GetLocalIP returns the local IP address that would be used for outbound connections.
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			return
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
