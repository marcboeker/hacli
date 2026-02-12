package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeServer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no_trailing_slash", "http://localhost:8123", "http://localhost:8123"},
		{"single_trailing_slash", "http://localhost:8123/", "http://localhost:8123"},
		{"multiple_trailing_slashes", "http://localhost:8123///", "http://localhost:8123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeServer(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseServerURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid_http", "http://localhost:8123", false},
		{"valid_https", "https://hass.example.com", false},
		{"with_trailing_slash", "http://localhost:8123/", false},
		{"empty", "", false}, // url.Parse accepts empty string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := ParseServerURL(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, url)
			}
		})
	}
}

func TestBuildAPIURL(t *testing.T) {
	tests := []struct {
		name     string
		server   string
		method   string
		expected string
		wantErr  bool
	}{
		{
			name:     "api_method_with_prefix",
			server:   "http://localhost:8123",
			method:   "/api/services",
			expected: "http://localhost:8123/api/services",
		},
		{
			name:     "api_method_without_prefix",
			server:   "http://localhost:8123",
			method:   "services",
			expected: "http://localhost:8123/api/services",
		},
		{
			name:     "api_method_already_has_api",
			server:   "http://localhost:8123",
			method:   "/api/states",
			expected: "http://localhost:8123/api/states",
		},
		{
			name:     "with_trailing_slash_in_server",
			server:   "http://localhost:8123/",
			method:   "states",
			expected: "http://localhost:8123/api/states",
		},
		{
			name:     "nested_path",
			server:   "http://localhost:8123",
			method:   "/api/services/light/turn_on",
			expected: "http://localhost:8123/api/services/light/turn_on",
		},
		{
			name:    "invalid_server",
			server:  "://invalid-url",
			method:  "states",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildAPIURL(tt.server, tt.method)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRESTCall_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/states", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"entity_id": "light.living_room", "state": "on"}]`))
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		Token:      "test_token",
		HTTPClient: server.Client(),
	}

	result, err := RESTCall(appCtx, "GET", "/api/states", nil)
	require.NoError(t, err)

	data, ok := result.([]any)
	require.True(t, ok)
	assert.Len(t, data, 1)
}

func TestRESTCall_Success_JSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		Token:      "test_token",
		HTTPClient: server.Client(),
	}

	result, err := RESTCall(appCtx, "POST", "services/light/turn_on", map[string]any{
		"entity_id": "light.living_room",
	})
	require.NoError(t, err)

	data, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "success", data["message"])
}

func TestRESTCall_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		HTTPClient: server.Client(),
	}

	result, err := RESTCall(appCtx, "POST", "services", nil)
	require.NoError(t, err)

	data, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Empty(t, data)
}

func TestRESTCall_NonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("plain text response"))
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		HTTPClient: server.Client(),
	}

	result, err := RESTCall(appCtx, "GET", "test", nil)
	require.NoError(t, err)

	str, ok := result.(string)
	require.True(t, ok)
	assert.Equal(t, "plain text response", str)
}

func TestRESTCall_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		HTTPClient: server.Client(),
	}

	_, err := RESTCall(appCtx, "GET", "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestRESTCall_InvalidServerURL(t *testing.T) {
	appCtx := &AppContext{
		Server:     "://invalid",
		HTTPClient: &http.Client{},
	}

	_, err := RESTCall(appCtx, "GET", "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server URL")
}

func TestRESTCall_InvalidPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not reach server with invalid payload")
	}))
	defer server.Close()

	appCtx := &AppContext{
		Server:     server.URL,
		HTTPClient: server.Client(),
	}

	// Try to marshal a channel (which cannot be marshaled)
	_, err := RESTCall(appCtx, "POST", "test", make(chan int))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshal payload")
}
