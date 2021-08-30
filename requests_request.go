package gorequests

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// doRead send request and read response
func (r *Request) doRead() error {
	return r.requestFactor(func() error {
		if err := r.doInternalRequest(); err != nil {
			return err
		}

		if r.isRead {
			return nil
		}
		r.isRead = true

		var err error
		r.bytes, err = ioutil.ReadAll(r.resp.Body)
		if err != nil {
			return errors.Wrapf(err, "read request(%s: %s) response failed", r.Method, r.cachedurl)
		}
		logger.Info(r.Context(), "[gorequests] %s: %s, doRead: %s", r.Method, r.cachedurl, r.bytes)
		return nil
	})
}

// doRequest send request
func (r *Request) doRequest() error {
	return r.requestFactor(r.doInternalRequest)
}

// doRequest send request
func (r *Request) doInternalRequest() error {
	if r.isRequest {
		return nil
	}

	r.cachedurl = r.parseRequestURL()
	r.isRequest = true

	logger.Info(r.Context(), "[gorequests] %s: %s", r.Method, r.cachedurl)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				_ = err // TODO: logs
			}
		}()
	}

	req, err := http.NewRequest(r.Method, r.cachedurl, r.Body)
	if err != nil {
		return errors.Wrapf(err, "new request(%s: %s) failed", r.Method, r.cachedurl)
	}

	req.Header = r.header

	// TODO: reuse client
	c := &http.Client{
		Timeout: r.Timeout,
	}
	if r.isIgnoreSSL {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	if r.persistentJar != nil {
		c.Jar = r.persistentJar
	}
	if r.isNoRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := c.Do(req)
	if err != nil {
		return errors.Wrapf(err, "do request(%s: %s) failed", r.Method, r.cachedurl)
	}
	r.resp = resp
	return nil
}

func (r *Request) requestFactor(f func() error) error {
	if r.err != nil {
		return r.err
	}

	r.reqlock.Lock()
	defer r.reqlock.Unlock()

	return f()
}
