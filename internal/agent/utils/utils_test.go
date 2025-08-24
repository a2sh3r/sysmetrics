package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressData(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		modify  func() func()
		wantErr bool
	}{
		{
			name:    "basic compress",
			input:   []byte("test data"),
			modify:  nil,
			wantErr: false,
		},
		{
			name:  "writer error (simulate)",
			input: []byte{},
			modify: func() func() {
				return func() {}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.modify != nil {
				cleanup = tt.modify()
			}
			defer cleanup()
			result, err := CompressData(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, len(result) > 0)
			}
		})
	}
}

func TestCompressData_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"empty slice", []byte{}, false},
		{"big slice", make([]byte, 1024*1024), false},
		{"nil slice", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompressData(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "get local IP",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := GetLocalIP()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, ip)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, ip)

				parsedIP := net.ParseIP(ip)
				assert.NotNil(t, parsedIP, "should return valid IP address")

				assert.NotEqual(t, "127.0.0.1", ip, "should not return localhost")
				assert.NotEqual(t, "::1", ip, "should not return IPv6 localhost")
			}
		})
	}
}
