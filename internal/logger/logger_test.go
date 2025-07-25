package logger

import (
	"testing"
	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name  string
		level string
		wantErr bool
	}{
		{"valid info", "info", false},
		{"valid debug", "debug", false},
		{"invalid", "notalevel", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Log)
			}
		})
	}
}

func TestAsyncLogging(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"AsyncInfo", func() { AsyncInfo("info", zap.String("k", "v")) }},
		{"AsyncWarn", func() { AsyncWarn("warn", zap.String("k", "v")) }},
		{"AsyncError", func() { AsyncError("err", zap.String("k", "v")) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.run)
		})
	}
} 