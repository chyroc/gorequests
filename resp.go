package gorequests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

func (r *Request) Unmarshal(val interface{}) error {
	bs, err := r.Bytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, val); err != nil {
		return fmt.Errorf("[gorequest] %s %s unmarshal %s to %s failed: %w", r.method, r.cachedurl, bs, reflect.TypeOf(val).Name(), err)
	}
	return nil
}

func (r *Request) MustUnmarshal(val interface{}) {
	_ = r.Unmarshal(val)
}

func (r *Request) Map() (map[string]interface{}, error) {
	bs, err := r.Bytes()
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, fmt.Errorf("[gorequest] %s %s unmarshal %s to map failed: %w", r.method, r.cachedurl, bs, err)
	}
	return m, nil
}

func (r *Request) MustMap() map[string]interface{} {
	val, _ := r.Map()
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
	val, _ := r.Text()
	return val
}

func (r *Request) Bytes() ([]byte, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}
	if err := r.doRead(); err != nil {
		return nil, err
	}

	return r.bytes, nil
}

func (r *Request) MustBytes() []byte {
	val, _ := r.Bytes()
	return val
}

func (r *Request) Response() (*http.Response, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp, nil
}

func (r *Request) MustResponse() *http.Response {
	val, _ := r.Response()
	return val
}

func (r *Request) ResponseStatus() (int, error) {
	if err := r.doRequest(); err != nil {
		return 0, err
	}

	return r.resp.StatusCode, nil
}

func (r *Request) MustResponseStatus() int {
	val, _ := r.ResponseStatus()
	return val
}

func (r *Request) ResponseHeaders() (http.Header, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp.Header, nil
}

func (r *Request) MustResponseHeaders() http.Header {
	val, _ := r.ResponseHeaders()
	return val
}

func (r *Request) ResponseHeadersByKey(key string) ([]string, error) {
	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp.Header.Values(key), nil
}

func (r *Request) MustResponseHeadersByKey(key string) []string {
	val, _ := r.ResponseHeadersByKey(key)
	return val
}

func (r *Request) MustResponseCookiesByKey(key string) []string {
	if err := r.doRequest(); err != nil {
		return nil
	}

	var resp []string
	for _, v := range r.resp.Cookies() {
		if v.Name == key {
			resp = append(resp, v.Value)
		}
	}
	return resp
}

func (r *Request) ResponseHeaderByKey(key string) (string, error) {
	if err := r.doRequest(); err != nil {
		return "", err
	}

	return r.resp.Header.Get(key), nil
}

func (r *Request) MustResponseHeaderByKey(key string) string {
	val, _ := r.ResponseHeaderByKey(key)
	return val
}
