// Package requestid provides a Traefik middleware plugin that injects
// an X-Request-ID header (UUIDv4) into HTTP requests and responses when
// the header is not already present.
package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Config holds the plugin configuration as deserialized from Traefik's
// dynamic configuration.
type Config struct {
	HeaderName string `json:"headerName,omitempty"`
	Enabled    bool   `json:"enabled,omitempty"`
}

// CreateConfig returns the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName: "X-Request-ID",
		Enabled:    true,
	}
}

// RequestID is the middleware handler.
type RequestID struct {
	next       http.Handler
	name       string
	headerName string
	enabled    bool
}

// New creates a new RequestID middleware.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &RequestID{
		next:       next,
		name:       name,
		headerName: config.HeaderName,
		enabled:    config.Enabled,
	}, nil
}

func (r *RequestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.enabled && req.Header.Get(r.headerName) == "" {
		id := uuid.Must(uuid.NewRandom()).String()
		req.Header.Set(r.headerName, id)
		rw.Header().Set(r.headerName, id)
	}

	r.next.ServeHTTP(rw, req)
}
