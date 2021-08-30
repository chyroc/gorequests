package gorequests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

// WithContext setup request context.Context
func (r *Request) WithContext(ctx context.Context) *Request {
	return r.configParamFactor(func(r *Request) {
		r.context = ctx
	})
}

// WithIgnoreSSL ignore ssl verify
func (r *Request) WithIgnoreSSL(ignore bool) *Request {
	return r.configParamFactor(func(r *Request) {
		r.isIgnoreSSL = ignore
	})
}

// WithHeader set one header k-v map
func (r *Request) WithHeader(k, v string) *Request {
	return r.configParamFactor(func(r *Request) {
		r.header.Add(k, v)
	})
}

// WithHeaders set multi header k-v map
func (r *Request) WithHeaders(kv map[string]string) *Request {
	return r.configParamFactor(func(r *Request) {
		for k, v := range kv {
			r.header.Add(k, v)
		}
	})
}

// WithRedirect set allow or not-allow redirect with Location header
func (r *Request) WithRedirect(b bool) *Request {
	return r.configParamFactor(func(r *Request) {
		r.isNoRedirect = !b
	})
}

// WithQuery set one query k-v map
func (r *Request) WithQuery(k, v string) *Request {
	return r.configParamFactor(func(r *Request) {
		r.querys[k] = append(r.querys[k], v)
	})
}

// WithQuerys set multi query k-v map
func (r *Request) WithQuerys(kv map[string]string) *Request {
	return r.configParamFactor(func(r *Request) {
		for k, v := range kv {
			r.querys[k] = append(r.querys[k], v)
		}
	})
}

// WithQueryStruct set multi query k-v map
func (r *Request) WithQueryStruct(v interface{}) *Request {
	return r.configParamFactor(func(r *Request) {
		kv, err := queryToMap(v)
		if err != nil {
			r.err = err
			return
		}
		for k, v := range kv {
			r.querys[k] = append(r.querys[k], v...)
		}
	})
}

// WithBody set request body, support: io.Reader, []byte, string, interface{}(as json format)
func (r *Request) WithBody(body interface{}) *Request {
	return r.configParamFactor(func(r *Request) {
		switch v := body.(type) {
		case io.Reader:
			r.Body = v
		case []byte:
			r.Body = bytes.NewReader(v)
		case string:
			r.Body = strings.NewReader(v)
		default:
			bs, err := json.Marshal(body)
			if err != nil {
				r.err = err
				return
			}
			r.Body = bytes.NewReader(bs)
		}
	})
}

// WithJSON set body same as WithBody, and set Content-Type to application/json
func (r *Request) WithJSON(body interface{}) *Request {
	return r.configParamFactor(func(r *Request) {
		r.WithBody(body)
		r.header.Set("Content-Type", "application/json")
	})
}

// WithForm set body and set Content-Type to multiform
func (r *Request) WithForm(body map[string]string) *Request {
	return r.configParamFactor(func(r *Request) {
		buf := bytes.Buffer{}
		f := multipart.NewWriter(&buf)
		for k, v := range body {
			if err := f.WriteField(k, v); err != nil {
				r.err = err
				return
			}
		}

		r.Body = strings.NewReader(buf.String())
		r.header.Set("Content-Type", f.FormDataContentType())
	})
}

// WithFile set file to body and set some multi-form k-v map
func (r *Request) WithFile(filename string, file io.Reader, fileKey string, params map[string]string) *Request {
	return r.configParamFactor(func(r *Request) {
		contentType, bod, err := newFileUploadRequest(params, fileKey, filename, file)
		if err != nil {
			r.err = err
			return
		}
		r.WithBody(bod)
		r.header.Set("Content-Type", contentType)
	})
}

// WithURLCookie set cookie of uri
func (r *Request) WithURLCookie(uri string) *Request {
	return r.configParamFactor(func(r *Request) {
		if r.persistentJar == nil {
			return
		}

		uriParse, err := url.Parse(uri)
		if err != nil {
			r.err = err
			return
		}

		cookies := []string{}
		for _, v := range r.persistentJar.Cookies(uriParse) {
			cookies = append(cookies, v.Name+"="+v.Value)
		}
		if len(cookies) > 0 {
			r.header.Add("cookie", strings.Join(cookies, "; ")) // use add not set
		}
	})
}

// WithHeader set one header k-v map
func (r *Request) configParamFactor(f func(*Request)) *Request {
	r.reqlock.Lock()
	defer r.reqlock.Unlock()

	if r.isRequest {
		r.SetError(fmt.Errorf("request %s %s alreday sended, cannot set request params", r.Method, r.cachedurl))
		return r
	}

	if r.err != nil {
		return r
	}

	f(r)

	return r
}
