// Package requestid provides a Traefik middleware plugin that injects
// an X-Request-ID header (UUIDv4) into HTTP requests and responses when
// the header is not already present.
package requestid

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
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
		id, err := newUUIDv4()
		if err == nil {
			req.Header.Set(r.headerName, id)
			rw.Header().Set(r.headerName, id)
		}
	}

	r.next.ServeHTTP(rw, req)
}

// newUUIDv4 generates a random UUID version 4 using crypto/rand.
func newUUIDv4() (string, error) {
	var b [16]byte

	_, err := rand.Read(b[:])
	if err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
