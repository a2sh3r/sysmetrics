package utils

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
)

func TestIsRetriableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"sql.ErrConnDone", sql.ErrConnDone, true},
		{"os.ErrDeadlineExceeded", os.ErrDeadlineExceeded, true},
		{"context.DeadlineExceeded", context.DeadlineExceeded, true},
		{"io.ErrUnexpectedEOF", io.ErrUnexpectedEOF, true},
		{"syscall.EAGAIN", syscall.EAGAIN, true},
		{"syscall.EWOULDBLOCK", syscall.EWOULDBLOCK, true},
		{"custom error", errors.New("custom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetriableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetriableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestWithRetries(t *testing.T) {
	type call struct {
		failCount int
		calls     int
	}
	tests := []struct {
		name      string
		failCount int
		wantErr   bool
	}{
		{"no error", 0, false},
		{"fail once, then success", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &call{failCount: tt.failCount}
			err := WithRetries(func() error {
				c.calls++
				if c.calls <= c.failCount {
					return sql.ErrConnDone
				}
				return nil
			})
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}
