package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/pkg/config"
)

type testConfig struct {
	App    testApp    `yaml:"app"`
	Server testServer `yaml:"server"`
}

type testApp struct {
	Environment string `yaml:"environment"`
}

type testServer struct {
	Port string `yaml:"port"`
}

func TestLoadYAML(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		wantErr  error
		validate func(t *testing.T, cfg *testConfig)
	}{
		{
			name: "success - valid yaml file",
			setup: func(t *testing.T) string {
				t.Helper()
				content := `app:
  environment: development
server:
  port: "8080"`
				return createTempFile(t, content)
			},
			wantErr: nil,
			validate: func(t *testing.T, cfg *testConfig) {
				t.Helper()
				assert.Equal(t, "development", cfg.App.Environment)
				assert.Equal(t, "8080", cfg.Server.Port)
			},
		},
		{
			name: "error - file does not exist",
			setup: func(t *testing.T) string {
				t.Helper()
				return "/nonexistent/path/config.yaml"
			},
			wantErr:  config.ErrReadFile,
			validate: nil,
		},
		{
			name: "error - invalid yaml syntax",
			setup: func(t *testing.T) string {
				t.Helper()
				content := `app: environment: development
server:
  port: "8080"`
				return createTempFile(t, content)
			},
			wantErr:  config.ErrUnmarshal,
			validate: nil,
		},
		{
			name: "success - empty yaml file",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, "")
			},
			wantErr: nil,
			validate: func(t *testing.T, cfg *testConfig) {
				t.Helper()
				assert.Empty(t, cfg.App.Environment)
				assert.Empty(t, cfg.Server.Port)
			},
		},
		{
			name: "success - partial yaml (missing fields)",
			setup: func(t *testing.T) string {
				t.Helper()
				content := `app:
  environment: production`
				return createTempFile(t, content)
			},
			wantErr: nil,
			validate: func(t *testing.T, cfg *testConfig) {
				t.Helper()
				assert.Equal(t, "production", cfg.App.Environment)
				assert.Empty(t, cfg.Server.Port)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := tt.setup(t)

			cfg, err := config.LoadYAML[testConfig](filepath)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadYAML_NestedStructs(t *testing.T) {
	type nestedConfig struct {
		Database struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"database"`
	}

	content := `database:
  host: localhost
  port: 5432
  username: admin
  password: secret123`

	filepath := createTempFile(t, content)
	cfg, err := config.LoadYAML[nestedConfig](filepath)

	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "admin", cfg.Database.Username)
	assert.Equal(t, "secret123", cfg.Database.Password)
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0o644)
	require.NoError(t, err)
	return tmpFile
}
