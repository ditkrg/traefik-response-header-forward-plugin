package traefik_response_header_forward_plugin

import (
	"context"
	"fmt"
	"net/http"
)

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
	next           http.Handler
	name           string
	requestHeaders []RequestHeader
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
		next:           next,
		name:           name,
		requestHeaders: config.RequestHeaders,
	}, nil
}

func (a *ResponseHeaderForward) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.next.ServeHTTP(rw, req)

	// for _, requestHeader := range a.requestHeaders {
	// 	headerValue := req.Header.Get(requestHeader.Name)
	// 	if headerValue == "" {
	// 		continue
	// 	}

	// 	rw.Header().Set(requestHeader.Name, headerValue)
	// }
}
