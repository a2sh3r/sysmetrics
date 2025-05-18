package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandler(t *testing.T) {
	repo := &mockRepo{}
	service := services.NewService(repo)

	handler := NewHandler(service, service, nil)
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.reader)
	assert.Equal(t, service, handler.writer)
}
