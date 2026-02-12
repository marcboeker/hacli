package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type contextKey struct{}

type ColumnDef struct {
	Header string
	Path   string
}

type AppContext struct {
	Server      string
	Token       string
	Password    string
	Timeout     int
	Output      string
	Verbose     int
	Insecure    bool
	Debug       bool
	Cert        string
	ColumnsRaw  string
	Columns     []string
	ColumnDefs  []ColumnDef
	NoHeaders   bool
	TableFormat string
	SortBy      string
	HTTPClient  *http.Client
	Stdout      io.Writer
	Stderr      io.Writer
}

func New() *AppContext {
	return &AppContext{
		Server:      envOr("HA_SERVER", "auto"),
		Token:       tokenFromEnv(),
		Password:    os.Getenv("HA_PASSWORD"),
		Timeout:     5,
		Output:      "auto",
		TableFormat: "markdown",
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

func tokenFromEnv() string {
	if token := os.Getenv("HA_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("HAIO_TOKEN")
}

func envOr(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func (ctx *AppContext) Resolve() error {
	if ctx.Server == "auto" || ctx.Server == "" {
		return fmt.Errorf("unable to automatically locate a Home Assistant server")
	}
	ctx.Columns = parseColumns(ctx.ColumnsRaw)

	transport := &http.Transport{}
	if ctx.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	if ctx.Cert != "" {
		certBytes, err := os.ReadFile(ctx.Cert)
		if err != nil {
			return fmt.Errorf("failed to read cert %q: %w", ctx.Cert, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(certBytes) {
			return fmt.Errorf("invalid certificate PEM in %q", ctx.Cert)
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.RootCAs = pool
	}

	ctx.HTTPClient = &http.Client{
		Timeout:   time.Duration(ctx.Timeout) * time.Second,
		Transport: transport,
	}

	return nil
}

func parseColumns(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	columns := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			columns = append(columns, trimmed)
		}
	}
	return columns
}

func (ctx *AppContext) AuthHeaders() map[string]string {
	headers := map[string]string{}
	if ctx.Token != "" {
		headers["Authorization"] = "Bearer " + ctx.Token
	} else if ctx.Password != "" {
		headers["x-ha-access"] = ctx.Password
	}
	return headers
}

func (ctx *AppContext) EffectiveOutput(defaultOutput string) string {
	if ctx.Output == "auto" {
		if defaultOutput == "data" {
			return "yaml"
		}
		if defaultOutput != "" {
			return defaultOutput
		}
		return "yaml"
	}
	return ctx.Output
}

func WithContext(parent context.Context, appCtx *AppContext) context.Context {
	return context.WithValue(parent, contextKey{}, appCtx)
}

func FromContext(ctx context.Context) *AppContext {
	if value := ctx.Value(contextKey{}); value != nil {
		if appCtx, ok := value.(*AppContext); ok {
			return appCtx
		}
	}
	return New()
}
