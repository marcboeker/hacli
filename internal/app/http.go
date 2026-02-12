package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/websocket"
)

func normalizeServer(server string) string {
	return strings.TrimRight(server, "/")
}

func ParseServerURL(server string) (*url.URL, error) {
	return url.Parse(normalizeServer(server))
}

func BuildAPIURL(server string, method string) (string, error) {
	base, err := url.Parse(normalizeServer(server))
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	path := strings.TrimPrefix(method, "/")
	if !strings.HasPrefix(path, "api/") {
		path = "api/" + path
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/" + path

	return base.String(), nil
}

func RESTCall(ctx *AppContext, verb string, method string, payload any) (any, error) {
	apiURL, err := BuildAPIURL(ctx.Server, method)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewBuffer(encoded)
	}

	req, err := http.NewRequest(strings.ToUpper(verb), apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for name, value := range ctx.AuthHeaders() {
		req.Header.Set(name, value)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := ctx.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("api error: %s", strings.TrimSpace(string(responseBody)))
	}

	if len(responseBody) == 0 {
		return map[string]any{}, nil
	}

	var decoded any
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return strings.TrimSpace(string(responseBody)), nil
	}

	return decoded, nil
}

func WSCall(ctx *AppContext, wsType string, wsData map[string]any) (any, error) {
	serverURL, err := url.Parse(normalizeServer(ctx.Server))
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	if serverURL.Scheme == "https" {
		serverURL.Scheme = "wss"
	} else {
		serverURL.Scheme = "ws"
	}
	serverURL.Path = strings.TrimRight(serverURL.Path, "/") + "/api/websocket"

	origin := "http://localhost/"
	conn, err := websocket.Dial(serverURL.String(), "", origin)
	if err != nil {
		return nil, fmt.Errorf("dial websocket: %w", err)
	}
	defer conn.Close()

	var greeting map[string]any
	if err := websocket.JSON.Receive(conn, &greeting); err != nil {
		return nil, fmt.Errorf("receive ws greeting: %w", err)
	}

	auth := map[string]any{"type": "auth"}
	if ctx.Token != "" {
		auth["access_token"] = ctx.Token
	}
	if err := websocket.JSON.Send(conn, auth); err != nil {
		return nil, fmt.Errorf("send ws auth: %w", err)
	}

	var authResp map[string]any
	if err := websocket.JSON.Receive(conn, &authResp); err != nil {
		return nil, fmt.Errorf("receive ws auth response: %w", err)
	}
	if authResp["type"] != "auth_ok" {
		return nil, fmt.Errorf("ws auth failed")
	}

	message := map[string]any{"id": 1, "type": wsType}
	for key, value := range wsData {
		message[key] = value
	}
	if err := websocket.JSON.Send(conn, message); err != nil {
		return nil, fmt.Errorf("send ws request: %w", err)
	}

	var result map[string]any
	if err := websocket.JSON.Receive(conn, &result); err != nil {
		return nil, fmt.Errorf("receive ws result: %w", err)
	}

	return result, nil
}
