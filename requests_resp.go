package gorequests

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

func (r *Request) Unmarshal(val interface{}) error {
	bs, err := r.Bytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, val); err != nil {
		return errors.Errorf("unmarshal %s to %s failed: %s", bs, reflect.TypeOf(val).Name(), err)
	}
	return nil
}

func (r *Request) MustUnmarshal(val interface{}) {
	err := r.Unmarshal(val)
	assert(err)
}

func (r *Request) Map() (map[string]interface{}, error) {
	bs, err := r.Bytes()
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp(%s) failed", bs)
	}
	return m, nil
}

func (r *Request) MustMap() map[string]interface{} {
	val, err := r.Map()
	assert(err)
	return val
}

func (r *Request) Text() (string, error) {
	bs, err := r.Bytes()
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (r *Request) MustText() string {
	val, err := r.Text()
	assert(err)
	return val
}

func (r *Request) Bytes() ([]byte, error) {
	if err := r.doRead(); err != nil {
		return nil, err
	}

	return r.bytes, nil
}

func (r *Request) MustBytes() []byte {
	val, err := r.Bytes()
	assert(err)
	return val
}

func (r *Request) Response() (*http.Response, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp, nil
}

func (r *Request) MustResponse() *http.Response {
	val, err := r.Response()
	assert(err)
	return val
}

func (r *Request) ResponseStatus() (int, error) {
	if err := r.doRequest(); err != nil {
		return 0, err
	}

	return r.resp.StatusCode, nil
}

func (r *Request) MustResponseStatus() int {
	val, err := r.ResponseStatus()
	assert(err)
	return val
}

func (r *Request) ResponseHeaders() (http.Header, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp.Header, nil
}

func (r *Request) MustResponseHeaders() http.Header {
	val, err := r.ResponseHeaders()
	assert(err)
	return val
}

func (r *Request) ResponseHeader(key string) ([]string, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	for k, v := range r.resp.Header {
		if key == k {
			return v, nil
		}
	}
	return nil, nil
}

func (r *Request) MustResponseHeader(key string) []string {
	val, err := r.ResponseHeader(key)
	assert(err)
	return val
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
