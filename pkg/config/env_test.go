package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/pkg/config"
)

func TestEnv_Required(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T)
		rules   []config.EnvKeyRule
		key     string
		wantErr error
		wantVal string
	}{
		{
			name: "success - required key - set via env var",
			key:  "TEST_REQUIRED_KEY",
			setup: func(t *testing.T) {
				t.Helper()
				t.Setenv("TEST_REQUIRED_KEY", "test-value")
			},
			rules:   nil,
			wantErr: nil,
			wantVal: "test-value",
		},
		{
			name: "failure - required key - not set",
			key:  "TEST_REQUIRED_MISSING",
			setup: func(t *testing.T) {
				t.Helper()
				// Ensure env var is not set
				os.Unsetenv("TEST_REQUIRED_MISSING")
			},
			rules:   nil,
			wantErr: config.ErrValidation,
			wantVal: "",
		},
		{
			name: "success - required key - with rules - valid",
			key:  "TEST_RULES_KEY",
			setup: func(t *testing.T) {
				t.Helper()
				t.Setenv("TEST_RULES_KEY", "valid-value")
			},
			rules:   []config.EnvKeyRule{config.MinLength(5)},
			wantErr: nil,
			wantVal: "valid-value",
		},
		{
			name: "failure - required key - with rules - invalid",
			key:  "TEST_RULES_INVALID",
			setup: func(t *testing.T) {
				t.Helper()
				t.Setenv("TEST_RULES_INVALID", "ab")
			},
			rules:   []config.EnvKeyRule{config.MinLength(5)},
			wantErr: config.ErrValidation,
			wantVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			env := config.NewEnv()
			if tt.rules != nil {
				env.Required(tt.key).WithRules(tt.rules...)
			} else {
				env.Required(tt.key)
			}
			env.WithEnvVars()

			err := env.Load()
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantVal, env.Get(tt.key))
		})
	}
}

func TestEnv_Optional(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T)
		getWithDefault bool
		key            string
		wantErr        error
		wantVal        string
	}{
		{
			name: "success - optional key - set",
			key:  "TEST_OPTIONAL_SET",
			setup: func(t *testing.T) {
				t.Helper()
				t.Setenv("TEST_OPTIONAL_SET", "optional-value")
			},
			getWithDefault: false,
			wantErr:        nil,
			wantVal:        "optional-value",
		},
		{
			name: "success - optional key - not set",
			key:  "TEST_OPTIONAL_NOT_SET",
			setup: func(t *testing.T) {
				t.Helper()
				os.Unsetenv("TEST_OPTIONAL_NOT_SET")
			},
			getWithDefault: false,
			wantErr:        nil,
			wantVal:        "",
		},
		{
			name: "success - optional key - with GetOrDefault",
			key:  "TEST_OPTIONAL_DEFAULT",
			setup: func(t *testing.T) {
				t.Helper()
				os.Unsetenv("TEST_OPTIONAL_DEFAULT")
			},
			getWithDefault: true,
			wantErr:        nil,
			wantVal:        "optional-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			env := config.NewEnv()
			env.Optional(tt.key)
			env.WithEnvVars()

			err := env.Load()
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if tt.getWithDefault {
				assert.Equal(
					t,
					"default-value",
					env.GetOrDefault("TEST_OPTIONAL_DEFAULT", "default-value"),
				)
			} else {
				assert.Equal(t, tt.wantVal, env.Get(tt.key))
			}
		})
	}
}

func TestEnv_IsPath(t *testing.T) {
	validPath := createTempEnvFile(t, "KEY=value")

	tests := []struct {
		name      string
		path      string
		mustExist bool
		want      bool
	}{
		{
			name:      "valid path - exists",
			path:      validPath,
			mustExist: true,
			want:      true,
		},
		{
			name:      "valid path - does not exist",
			path:      "/nonexistent/path.yaml",
			mustExist: true,
			want:      false,
		},
		{
			name:      "valid path format - does not need to exist",
			path:      "some/relative/path.yaml",
			mustExist: false,
			want:      true,
		},
		{
			name:      "invalid path format",
			path:      "/nonexistent(/path.yam;",
			mustExist: true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := config.IsPath(tt.mustExist)(tt.path)
			assert.Equal(t, tt.want, isValid)
		})
	}
}

func TestEnv_EnvFiles(t *testing.T) {
	t.Run("load from env file", func(t *testing.T) {
		envContent := `TEST_FILE_KEY=file-value
TEST_FILE_KEY2=file-value2`
		envFile := createTempEnvFile(t, envContent)

		env := config.NewEnv()
		env.Required("TEST_FILE_KEY")
		env.Required("TEST_FILE_KEY2")
		env.AddEnvFiles(envFile)

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "file-value", env.Get("TEST_FILE_KEY"))
		assert.Equal(t, "file-value2", env.Get("TEST_FILE_KEY2"))
	})

	t.Run("env file not found", func(t *testing.T) {
		env := config.NewEnv()
		env.Required("TEST_KEY")
		env.AddEnvFiles("/nonexistent/path/.env")

		err := env.Load()
		// Should still validate even if file doesn't exist
		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrValidation)
	})

	t.Run("multiple env files - first wins", func(t *testing.T) {
		envContent1 := `TEST_MULTI=first`
		envContent2 := `TEST_MULTI=second`

		file1 := createTempEnvFile(t, envContent1)
		file2 := createTempEnvFile(t, envContent2)

		env := config.NewEnv()
		env.Required("TEST_MULTI")
		env.AddEnvFiles(file1, file2)

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "first", env.Get("TEST_MULTI"))
	})
}

func TestEnv_EnvVarsPriority(t *testing.T) {
	// Env vars should take priority over env files
	t.Run("env var priority over file", func(t *testing.T) {
		t.Setenv("TEST_PRIORITY", "from-env-var")
		defer os.Unsetenv("TEST_PRIORITY")

		envContent := `TEST_PRIORITY=from-file`
		envFile := createTempEnvFile(t, envContent)

		env := config.NewEnv()
		env.Required("TEST_PRIORITY")
		env.WithEnvVars()
		env.AddEnvFiles(envFile)

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "from-env-var", env.Get("TEST_PRIORITY"))
	})
}

func TestEnv_MixedSources(t *testing.T) {
	t.Run("mix of env vars and files", func(t *testing.T) {
		t.Setenv("TEST_MIX_ENV", "env-value")
		defer os.Unsetenv("TEST_MIX_ENV")

		envContent := `TEST_MIX_FILE=file-value`
		envFile := createTempEnvFile(t, envContent)

		env := config.NewEnv()
		env.Required("TEST_MIX_ENV")
		env.Required("TEST_MIX_FILE")
		env.WithEnvVars()
		env.AddEnvFiles(envFile)

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "env-value", env.Get("TEST_MIX_ENV"))
		assert.Equal(t, "file-value", env.Get("TEST_MIX_FILE"))
	})
}

func TestEnv_NoEnvVars(t *testing.T) {
	t.Run("only files, no env vars", func(t *testing.T) {
		envContent := `TEST_ONLY_FILE=from-file`
		envFile := createTempEnvFile(t, envContent)

		env := config.NewEnv()
		env.Required("TEST_ONLY_FILE")
		env.AddEnvFiles(envFile)

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "from-file", env.Get("TEST_ONLY_FILE"))
	})
}

func TestEnv_NoFiles(t *testing.T) {
	t.Run("only env vars, no files", func(t *testing.T) {
		t.Setenv("TEST_ONLY_ENV", "from-env")

		env := config.NewEnv()
		env.Required("TEST_ONLY_ENV")
		env.WithEnvVars()

		err := env.Load()
		require.NoError(t, err)
		assert.Equal(t, "from-env", env.Get("TEST_ONLY_ENV"))
	})
}

func createTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(tmpFile, []byte(content), 0o644)
	require.NoError(t, err)
	return tmpFile
}
