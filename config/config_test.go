package config

import (
	"os"
	"testing"

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
			name: "ValidConfig",
			envVars: map[string]string{
				"GRPC_PORT":         "50051",
				"GRPC_GATEWAY_PORT": "8080",
				"POSTGRES_HOST":     "localhost",
				"POSTGRES_PORT":     "5432",
				"POSTGRES_DB":       "testdb",
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpassword",
				"POSTGRES_MAX_CONN": "10",
			},
			wantConfig: &Config{
				GRPC: GRPC{
					Port:        "50051",
					GatewayPort: "8080",
				},
				PG: PG{
					URL:      "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable&pool_max_conns=10",
					Host:     "localhost",
					Port:     "5432",
					DB:       "testdb",
					User:     "testuser",
					Password: "testpassword",
					MaxConn:  "10",
				},
			},
			wantErr: false,
		},
		{
			name: "MissingEnvVars",
			envVars: map[string]string{
				"GRPC_PORT":         "",
				"GRPC_GATEWAY_PORT": "",
				"POSTGRES_HOST":     "",
				"POSTGRES_PORT":     "",
				"POSTGRES_DB":       "",
				"POSTGRES_USER":     "",
				"POSTGRES_PASSWORD": "",
				"POSTGRES_MAX_CONN": "",
			},
			wantConfig: &Config{
				GRPC: GRPC{
					Port:        "9090",
					GatewayPort: "8080",
				},
				PG: PG{
					URL:      "postgres://user:1234567@localhost:5432/library?sslmode=disable&pool_max_conns=10",
					Host:     "localhost",
					Port:     "5432",
					DB:       "library",
					User:     "user",
					Password: "1234567",
					MaxConn:  "10",
				},
			},
			wantErr: false,
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
