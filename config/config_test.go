package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:parallel // explanation
func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		wantConfig *Config
		wantErr    bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"GRPC_PORT":                 "50051",
				"GRPC_GATEWAY_PORT":         "8080",
				"POSTGRES_HOST":             "localhost",
				"POSTGRES_PORT":             "5432",
				"POSTGRES_DB":               "testdb",
				"POSTGRES_USER":             "user",
				"POSTGRES_PASSWORD":         "password",
				"POSTGRES_MAX_CONN":         "10",
				"OUTBOX_ENABLED":            "true",
				"OUTBOX_WORKERS":            "5",
				"OUTBOX_BATCH_SIZE":         "100",
				"OUTBOX_WAIT_TIME_MS":       "500",
				"OUTBOX_IN_PROGRESS_TTL_MS": "1000",
				"OUTBOX_BOOK_SEND_URL":      "http://book-service/send",
				"OUTBOX_AUTHOR_SEND_URL":    "http://author-service/send",
			},
			wantConfig: &Config{
				GRPC: GRPC{
					Port:        "50051",
					GatewayPort: "8080",
				},
				PG: PG{
					Host:     "localhost",
					Port:     "5432",
					DB:       "testdb",
					User:     "user",
					Password: "password",
					MaxConn:  "10",
					URL:      "postgres://user:password@localhost:5432/testdb?sslmode=disable&pool_max_conns=10",
				},
				Outbox: Outbox{
					Enabled:         true,
					Workers:         5,
					BatchSize:       100,
					WaitTimeMS:      500 * time.Millisecond,
					InProgressTTLMS: 1000 * time.Millisecond,
					BookSendURL:     "http://book-service/send",
					AuthorSendURL:   "http://author-service/send",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid outbox enabled",
			envVars: map[string]string{
				"OUTBOX_ENABLED": "invalid",
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid outbox workers",
			envVars: map[string]string{
				"OUTBOX_ENABLED": "true",
				"OUTBOX_WORKERS": "invalid workers",
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid outbox workers",
			envVars: map[string]string{
				"OUTBOX_ENABLED":    "true",
				"OUTBOX_WORKERS":    "5",
				"OUTBOX_BATCH_SIZE": "invalid batch size",
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid outbox wait time",
			envVars: map[string]string{
				"OUTBOX_ENABLED":      "true",
				"OUTBOX_WORKERS":      "5",
				"OUTBOX_BATCH_SIZE":   "100",
				"OUTBOX_WAIT_TIME_MS": "invalid wait time",
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid outbox progress TTL",
			envVars: map[string]string{
				"OUTBOX_ENABLED":            "true",
				"OUTBOX_WORKERS":            "5",
				"OUTBOX_BATCH_SIZE":         "100",
				"OUTBOX_WAIT_TIME_MS":       "1000",
				"OUTBOX_IN_PROGRESS_TTL_MS": "invalid progress TTl",
			},
			wantConfig: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := New()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantConfig, cfg)
			}

			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
