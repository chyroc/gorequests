package gorequests

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Context request context.Context
func (r *Request) Context() context.Context {
	if r.context != nil {
		return r.context
	}
	return context.TODO()
}

// Timeout request timeout
func (r *Request) Timeout() time.Duration {
	return r.timeout
}

// URL request url
func (r *Request) URL() string {
	return r.url
}

// RequestFullURL request full url, contain query param
func (r *Request) RequestFullURL() string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.parseRequestURL()
}

// Method request method
func (r *Request) Method() string {
	return r.method
}

// RequestHeader request header
func (r *Request) RequestHeader() http.Header {
	return r.header
}

// request url
func (r *Request) parseRequestURL() string {
	URL, err := url.Parse(r.url)
	if err != nil {
		return r.url
	}
	q := URL.Query()
	for k, v := range r.querys {
		q[k] = append(q[k], v...)
	}
	URL.RawQuery = q.Encode()
	return URL.String()
}
