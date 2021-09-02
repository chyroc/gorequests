package gorequests

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

// doRead send request and read response
func (r *Request) doRead() error {
	return r.requestFactor(func() error {
		if err := r.doInternalRequest(); err != nil {
			r.err = err
			return err
		}

		if r.isRead {
			return nil
		}
		r.isRead = true

		var err error
		fmt.Println("r.resp is nil", r.resp == nil)
		r.bytes, err = ioutil.ReadAll(r.resp.Body)
		if err != nil {
			return fmt.Errorf("[gorequest] %s %s read response failed: %w", r.method, r.cachedurl, err)
		}

		r.logger.Info(r.Context(), "[gorequests] %s: %s, doRead: %s", r.method, r.cachedurl, r.bytes)
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

	r.logger.Info(r.Context(), "[gorequests] %s: %s", r.method, r.cachedurl)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				r.logger.Error(r.Context(), "save cookie failed: %s", err)
			}
		}()
	}

	req, err := http.NewRequest(r.method, r.cachedurl, r.body)
	if err != nil {
		return fmt.Errorf("[gorequest] %s %s new request failed: %w", r.method, r.cachedurl, err)
	}

	req.Header = r.header

	// TODO: reuse client
	c := &http.Client{
		Timeout: r.timeout,
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

	r.isRequest = true
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("[gorequest] %s %s send request failed: %w", r.method, r.cachedurl, err)
	}
	r.resp = resp
	return nil
}

func (r *Request) requestFactor(f func() error) error {
	if r.err != nil {
		return r.err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.err = f()
	return r.err
}
