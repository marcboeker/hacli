package app

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	appCtx := New()
	require.NotNil(t, appCtx)

	assert.Equal(t, "auto", appCtx.Server)
	assert.Equal(t, 5, appCtx.Timeout)
	assert.Equal(t, "auto", appCtx.Output)
	assert.Equal(t, "markdown", appCtx.TableFormat)
	assert.Equal(t, os.Stdout, appCtx.Stdout)
	assert.Equal(t, os.Stderr, appCtx.Stderr)
}

func TestEnvOr_WithValue(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := envOr("TEST_VAR", "fallback")
	assert.Equal(t, "test_value", result)
}

func TestEnvOr_WithEmptyValue(t *testing.T) {
	os.Setenv("TEST_VAR", "")
	defer os.Unsetenv("TEST_VAR")

	result := envOr("TEST_VAR", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestEnvOr_WithoutValue(t *testing.T) {
	os.Unsetenv("TEST_VAR_NOT_SET")

	result := envOr("TEST_VAR_NOT_SET", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestTokenFromEnv_HA_TOKEN(t *testing.T) {
	os.Setenv("HA_TOKEN", "my_token")
	defer os.Unsetenv("HA_TOKEN")
	os.Unsetenv("HAIO_TOKEN")

	result := tokenFromEnv()
	assert.Equal(t, "my_token", result)
}

func TestTokenFromEnv_HAIO_TOKEN(t *testing.T) {
	os.Unsetenv("HA_TOKEN")
	os.Setenv("HAIO_TOKEN", "my_hasio_token")
	defer os.Unsetenv("HAIO_TOKEN")

	result := tokenFromEnv()
	assert.Equal(t, "my_hasio_token", result)
}

func TestParseColumns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"single", "name", []string{"name"}},
		{"multiple", "name,state,entity_id", []string{"name", "state", "entity_id"}},
		{"with_spaces", "  name  ,  state  ", []string{"name", "state"}},
		{"empty_parts", "name,,state", []string{"name", "state"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseColumns(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppContext_Resolve_AutoServer(t *testing.T) {
	appCtx := New()
	appCtx.Server = "auto"

	err := appCtx.Resolve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to automatically locate")
}

func TestAppContext_Resolve_EmptyServer(t *testing.T) {
	appCtx := New()
	appCtx.Server = ""

	err := appCtx.Resolve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to automatically locate")
}

func TestAppContext_Resolve_WithServer(t *testing.T) {
	appCtx := New()
	appCtx.Server = "http://localhost:8123"

	err := appCtx.Resolve()
	assert.NoError(t, err)
	assert.NotNil(t, appCtx.HTTPClient)
}

func TestAppContext_AuthHeaders_Token(t *testing.T) {
	appCtx := &AppContext{Token: "my_token"}
	headers := appCtx.AuthHeaders()

	assert.Equal(t, "Bearer my_token", headers["Authorization"])
	assert.Empty(t, headers["x-ha-access"])
}

func TestAppContext_AuthHeaders_Password(t *testing.T) {
	appCtx := &AppContext{Password: "my_password"}
	headers := appCtx.AuthHeaders()

	assert.Equal(t, "my_password", headers["x-ha-access"])
	assert.Empty(t, headers["Authorization"])
}

func TestAppContext_AuthHeaders_Empty(t *testing.T) {
	appCtx := &AppContext{}
	headers := appCtx.AuthHeaders()

	assert.Empty(t, headers)
}

func TestAppContext_EffectiveOutput(t *testing.T) {
	tests := []struct {
		name          string
		ctxOutput     string
		defaultOutput string
		expected      string
	}{
		{"auto_with_data", "auto", "data", "yaml"},
		{"auto_with_empty", "auto", "", "yaml"},
		{"auto_with_json", "auto", "json", "json"},
		{"explicit_json", "json", "yaml", "json"},
		{"explicit_yaml", "yaml", "json", "yaml"},
		{"explicit_table", "table", "yaml", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCtx := &AppContext{Output: tt.ctxOutput}
			result := appCtx.EffectiveOutput(tt.defaultOutput)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithContext_FromContext(t *testing.T) {
	appCtx := &AppContext{Server: "http://test"}
	ctx := WithContext(context.Background(), appCtx)

	retrieved := FromContext(ctx)
	assert.Equal(t, appCtx, retrieved)
}

func TestFromContext_NotSet(t *testing.T) {
	ctx := context.Background()

	retrieved := FromContext(ctx)
	assert.NotNil(t, retrieved)
	// Should return a new default context
	assert.Equal(t, "auto", retrieved.Server)
}

func TestFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey{}, "not an AppContext")

	retrieved := FromContext(ctx)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "auto", retrieved.Server)
}

func TestAppContext_Resolve_WithCert_InvalidPath(t *testing.T) {
	appCtx := New()
	appCtx.Server = "http://localhost:8123"
	appCtx.Cert = "/nonexistent/cert.pem"

	err := appCtx.Resolve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read cert")
}

func TestAppContext_Resolve_WithCert_InvalidPEM(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_cert_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("not a valid pem")
	require.NoError(t, err)
	tmpFile.Close()

	appCtx := New()
	appCtx.Server = "http://localhost:8123"
	appCtx.Cert = tmpFile.Name()

	err = appCtx.Resolve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid certificate PEM")
}
