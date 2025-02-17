package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadDefaults(t *testing.T) {
	t.Run("load defaults", func(t *testing.T) {
		var c Config
		c.LoadDefaults()

		assert.Equal(t, "0.0.0.0", c.Host)
		assert.Equal(t, uint16(8080), c.Port)
		assert.Equal(t, 65536, c.MaxBody)
		assert.Equal(t, 10*time.Second, c.ReadTimeout)
		assert.Equal(t, 10*time.Second, c.WriteTimeout)
		assert.Equal(t, 5*time.Minute, c.IdleTimeout)
	})
}

func TestConfig_LoadOptions(t *testing.T) {
	testCases := []struct {
		name         string
		opts         []Option
		wantMTLS     bool
		wantHost     string
		wantPort     uint16
		wantName     string
		wantMax      int
		wantCA       string
		wantCert     string
		wantKey      string
		wantRead     time.Duration
		wantWrite    time.Duration
		wantIdle     time.Duration
		wantRecover  bool
		wantSecurity bool
		wantOrigins  string
		wantHeaders  string
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			opts: []Option{},
		},
		{
			name:     "mTLS",
			opts:     []Option{WithMutualTLS()},
			wantMTLS: true,
		},
		{
			name:     "host&port",
			opts:     []Option{WithHostPort("localhost", 8080)},
			wantHost: "localhost",
			wantPort: 8080,
		},
		{
			name:     "app name",
			opts:     []Option{WithAppName("My Fantastic App")},
			wantName: "My Fantastic App",
		},
		{
			name:    "max body",
			opts:    []Option{WithMaxBody(12345)},
			wantMax: 12345,
		},
		{
			name:     "tls",
			opts:     []Option{WithTLS("file1", "file2", "file3")},
			wantCA:   "file1",
			wantCert: "file2",
			wantKey:  "file3",
		},
		{
			name:      "timeouts",
			opts:      []Option{WithTimeouts(time.Second*4, time.Second*8, time.Minute*9)},
			wantRead:  time.Second * 4,
			wantWrite: time.Second * 8,
			wantIdle:  time.Minute * 9,
		},
		{
			name:        "panic recovery",
			opts:        []Option{WithRecovery()},
			wantRecover: true,
		},
		{
			name:         "high security",
			opts:         []Option{WithSecurity()},
			wantSecurity: true,
		},
		{
			name:        "cors",
			opts:        []Option{WithCORS("*", "Cache-Control")},
			wantOrigins: "*",
			wantHeaders: "Cache-Control",
		},
		{
			name: "all",
			opts: []Option{
				WithMutualTLS(),
				WithHostPort("localhost", 8080),
				WithAppName("My Fantastic App"),
				WithMaxBody(12345),
				WithTLS("file1", "file2", "file3"),
				WithTimeouts(time.Second*4, time.Second*8, time.Minute*9),
				WithRecovery(),
				WithSecurity(),
				WithCORS("*", "Cache-Control"),
			},
			wantMTLS:     true,
			wantHost:     "localhost",
			wantPort:     8080,
			wantName:     "My Fantastic App",
			wantMax:      12345,
			wantCA:       "file1",
			wantCert:     "file2",
			wantKey:      "file3",
			wantRead:     time.Second * 4,
			wantWrite:    time.Second * 8,
			wantIdle:     time.Minute * 9,
			wantRecover:  true,
			wantSecurity: true,
			wantOrigins:  "*",
			wantHeaders:  "Cache-Control",
		},
		{
			name:      "defaults",
			opts:      []Option{WithDefaults()},
			wantHost:  "0.0.0.0",
			wantPort:  8080,
			wantMax:   65536,
			wantRead:  10 * time.Second,
			wantWrite: 10 * time.Second,
			wantIdle:  5 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			c.LoadOptions(tc.opts...)

			assert.Equal(t, tc.wantMTLS, c.MutualTLS)
			assert.Equal(t, tc.wantHost, c.Host)
			assert.Equal(t, tc.wantPort, c.Port)
			assert.Equal(t, tc.wantName, c.AppName)
			assert.Equal(t, tc.wantMax, c.MaxBody)
			assert.Equal(t, tc.wantRead, c.ReadTimeout)
			assert.Equal(t, tc.wantWrite, c.WriteTimeout)
			assert.Equal(t, tc.wantIdle, c.IdleTimeout)
			assert.Equal(t, tc.wantCA, c.CA)
			assert.Equal(t, tc.wantCert, c.TLSCert)
			assert.Equal(t, tc.wantKey, c.TLSKey)
			assert.Equal(t, tc.wantRecover, c.RecoverPanics)
			assert.Equal(t, tc.wantSecurity, c.HighSecurity)
			assert.Equal(t, tc.wantOrigins, c.CorsOrigins)
			assert.Equal(t, tc.wantHeaders, c.CorsHeaders)

		})
	}
}
