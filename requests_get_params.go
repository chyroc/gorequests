package gorequests

import (
	"context"
	"net/http"
	"net/url"
)

// RequestURL get request request url
func (r *Request) RequestURL() string {
	r.reqlock.RLock()
	defer r.reqlock.RUnlock()

	return r.parseRequestURL()
}

// Context request context.Context
func (r *Request) Context() context.Context {
	if r.context != nil {
		return r.context
	}
	return context.TODO()
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
