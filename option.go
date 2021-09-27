package gorequests

import (
	"time"
)

type RequestOption func(req *Request) error

func WithLogger(logger Logger) RequestOption {
	return func(req *Request) error {
		req.WithLogger(logger)
		return nil
	}
}

func WithTimeout(timeout time.Duration) RequestOption {
	return func(req *Request) error {
		req.WithTimeout(timeout)
		return nil
	}
}

func WithHeader(key, val string) RequestOption {
	return func(req *Request) error {
		req.WithHeader(key, val)
		return nil
	}
}

func WithQuery(key, val string) RequestOption {
	return func(req *Request) error {
		req.WithQuery(key, val)
		return nil
	}
}
