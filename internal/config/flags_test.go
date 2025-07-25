package config

import (
	"flag"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNetAddress_String(t *testing.T) {
	tests := []struct {
		name string
		addr NetAddress
		want string
	}{
		{"host:port", NetAddress{"localhost", 8080}, "localhost:8080"},
		{"empty host", NetAddress{"", 1234}, ":1234"},
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
		wantPoll  float64
		wantRep   float64
		wantKey   string
		wantLimit int64
	}{
		{
			name:     "all flags",
			args:     []string{"cmd", "-a", "localhost:8081", "-p", "5", "-r", "15", "-k", "secret", "-l", "3"},
			wantAddr: "http://localhost:8081",
			wantPoll: 5,
			wantRep:  15,
			wantKey:  "secret",
			wantLimit: 3,
		},
		{
			name:     "defaults",
			args:     []string{"cmd"},
			wantAddr: "",
			wantPoll: 2,
			wantRep:  10,
			wantKey:  "",
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