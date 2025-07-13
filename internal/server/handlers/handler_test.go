package handlers

import (
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	repo := &mockRepo{}
	service := services.NewService(repo)

	handler := NewHandler(service, service, nil)
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.reader)
	assert.Equal(t, service, handler.writer)
}

func BenchmarkNewHandler(b *testing.B) {
	repo := &mockRepo{}
	service := services.NewService(repo)
	for i := 0; i < b.N; i++ {
		_ = NewHandler(service, service, nil)
	}
}
