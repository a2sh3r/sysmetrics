package hash

import (
	"testing"
)

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		key      string
		expected string
	}{
		{
			name:     "Test #1 empty key",
			data:     "test data",
			key:      "",
			expected: "",
		},
		{
			name:     "Test #2 valid hash",
			data:     "test data",
			key:      "test key",
			expected: "7d4e3fec20c39323f1a3f1732e29f653cded4b1066b5e2e0a9f3c7c9e8c1b2a3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := CalculateHash(tt.data, tt.key)
			if tt.key == "" {
				if hash != "" {
					t.Errorf("CalculateHash() = %v, want empty string", hash)
				}
			} else {
				if hash == "" {
					t.Error("CalculateHash() returned empty string for non-empty key")
				}
				if len(hash) != 64 {
					t.Errorf("CalculateHash() returned hash of length %d, want 64", len(hash))
				}
			}
		})
	}
}

func TestVerifyHash(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		key     string
		hash    string
		wantErr bool
	}{
		{
			name:    "Test #1 empty key",
			data:    "test data",
			key:     "",
			hash:    "any hash",
			wantErr: false,
		},
		{
			name:    "Test #2 valid hash",
			data:    "test data",
			key:     "test key",
			hash:    CalculateHash("test data", "test key"),
			wantErr: false,
		},
		{
			name:    "Test #3 invalid hash",
			data:    "test data",
			key:     "test key",
			hash:    "invalid hash",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyHash(tt.data, tt.key, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
