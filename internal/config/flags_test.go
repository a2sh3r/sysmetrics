package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNetAddress_String(t *testing.T) {
	tests := []struct {
		name string
		addr NetAddress
		want string
	}{
		{"host:port", NetAddress{Port: 8080, Host: "localhost"}, "localhost:8080"},
		{"empty host", NetAddress{Port: 1234, Host: ""}, ":1234"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNetAddress_Set(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHost  string
		wantPort  int
		wantError bool
	}{
		{"valid address", "localhost:8080", "localhost", 8080, false},
		{"invalid format", "localhost", "", 0, true},
		{"invalid port", "localhost:abc", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addr NetAddress
			err := addr.Set(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHost, addr.Host)
				assert.Equal(t, tt.wantPort, addr.Port)
			}
		})
	}
}

func TestAgentConfig_ParseFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantAddr  string
		wantPoll  Duration
		wantRep   Duration
		wantKey   string
		wantLimit int64
	}{
		{
			name:      "all flags",
			args:      []string{"cmd", "-a", "localhost:8081", "-p", "5s", "-r", "15s", "-k", "secret", "-l", "3"},
			wantAddr:  "http://localhost:8081",
			wantPoll:  Duration{5 * time.Second},
			wantRep:   Duration{15 * time.Second},
			wantKey:   "secret",
			wantLimit: 3,
		},
		{
			name:      "defaults",
			args:      []string{"cmd"},
			wantAddr:  "",
			wantPoll:  Duration{2 * time.Second},
			wantRep:   Duration{10 * time.Second},
			wantKey:   "",
			wantLimit: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ExitOnError)
			cfg := &AgentConfig{}
			os.Args = tt.args
			cfg.ParseFlags()
			assert.Equal(t, tt.wantAddr, cfg.Address)
			assert.Equal(t, tt.wantPoll, cfg.PollInterval)
			assert.Equal(t, tt.wantRep, cfg.ReportInterval)
			assert.Equal(t, tt.wantKey, cfg.SecretKey)
			assert.Equal(t, tt.wantLimit, cfg.RateLimit)
		})
	}
}
