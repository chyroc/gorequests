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
