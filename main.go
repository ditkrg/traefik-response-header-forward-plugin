package traefik_response_header_forward_plugin

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

var (
	_ interface {
		http.ResponseWriter
		http.Hijacker
	} = &wrappedResponseWriter{}
)

type wrappedResponseWriter struct {
	rw   http.ResponseWriter
	buf  *bytes.Buffer
	code int
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.rw.Header()
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *wrappedResponseWriter) Flush() {
	w.rw.WriteHeader(w.code)
	io.Copy(w.rw, w.buf)
}

func (w *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.rw.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not an http.Hijacker", w.rw)
	}

	return hijacker.Hijack()
}

// ========================================

type RequestHeader struct {
	Name string `json:"name,omitempty"`
}

type Config struct {
	RequestHeaders []RequestHeader `json:"requestHeaders,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RequestHeaders: make([]RequestHeader, 0),
	}
}

type ResponseHeaderForward struct {
	next   http.Handler
	name   string
	config *Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.RequestHeaders) == 0 {
		return nil, fmt.Errorf("RequestHeaders cannot be empty")

	}

	for _, headerName := range config.RequestHeaders {
		if headerName.Name == "" {
			return nil, fmt.Errorf("RequestHeaders.Name cannot be empty")
		}
	}
	return &ResponseHeaderForward{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

func (a *ResponseHeaderForward) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	resp := &wrappedResponseWriter{
		rw:  rw,
		buf: &bytes.Buffer{},
	}

	defer resp.Flush()

	a.next.ServeHTTP(resp, req)

	for _, requestHeader := range a.config.RequestHeaders {
		headerValue := req.Header.Get(requestHeader.Name)
		if headerValue == "" {
			continue
		}

		resp.Header().Set(requestHeader.Name, headerValue)
	}
}
