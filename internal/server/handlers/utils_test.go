package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name string
		params []string
		wantErr bool
	}{
		{"all valid", []string{"gauge", "name", "123"}, false},
		{"empty param", []string{"gauge", "", "123"}, true},
		{"all empty", []string{"", "", ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParams(tt.params...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 