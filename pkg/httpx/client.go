package httpx

import (
	"crypto/tls"
	"net/http"
	"strings"
)

type CustomTransport struct {
	http.RoundTripper
}

func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Scheme, "https") {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return transport.RoundTrip(req)
	}

	// For HTTP, use the default transport
	return http.DefaultTransport.RoundTrip(req)
}

type MyCustomTransport struct {
	httpTransport *http.Transport
}

func NewMyCustomTransport() *MyCustomTransport {
	return &MyCustomTransport{
		httpTransport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (t *MyCustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Scheme, "https") {
		return t.httpTransport.RoundTrip(req)
	}

	// For HTTP, use the default transport
	return http.DefaultTransport.RoundTrip(req)
}
