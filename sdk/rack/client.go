package rack

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"
)

type Client struct {
	Host    string
	Key     string
	Rack    string
	Socket  string
	Version string
}

type Params map[string]string

func (c *Client) GetStream(path string) (io.ReadCloser, error) {
	req, err := c.Request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.handleRequest(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) Get(path string, out interface{}) error {
	r, err := c.GetStream(path)
	if err != nil {
		return err
	}

	return unmarshalReader(r, out)
}

func (c *Client) PostStream(path string, body io.Reader) (io.ReadCloser, error) {
	req, err := c.Request("POST", path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.handleRequest(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) Post(path string, params Params, out interface{}) error {
	uv := url.Values{}

	for k, v := range params {
		uv.Set(k, v)
	}

	r, err := c.PostStream(path, bytes.NewReader([]byte(uv.Encode())))
	if err != nil {
		return err
	}

	return unmarshalReader(r, out)
}

func (c *Client) PutStream(path string, body io.Reader) (io.ReadCloser, error) {
	req, err := c.Request("PUT", path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.handleRequest(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) Put(path string, params Params, out interface{}) error {
	uv := url.Values{}

	for k, v := range params {
		uv.Set(k, v)
	}

	r, err := c.PutStream(path, bytes.NewReader([]byte(uv.Encode())))
	if err != nil {
		return err
	}

	return unmarshalReader(r, out)
}

func (c *Client) Delete(path string, out interface{}) error {
	req, err := c.Request("DELETE", path, nil)
	if err != nil {
		return err
	}

	res, err := c.handleRequest(req)
	if err != nil {
		return err
	}

	return unmarshalReader(res.Body, out)
}

func (c *Client) Client() *http.Client {
	t := &http.Transport{
		DialContext: func(ctx context.Context, proto, addr string) (net.Conn, error) {
			if c.Socket != "" {
				return (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext(ctx, "unix", c.Socket)
			}

			return (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext(ctx, proto, addr)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if err := http2.ConfigureTransport(t); err != nil {
		panic(err)
	}

	return &http.Client{
		Transport: t,
	}
}

func (c *Client) Request(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://%s%s", c.Host, path), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Rack", c.Rack)
	req.Header.Add("Version", c.Version)

	req.SetBasicAuth("convox", string(c.Key))

	return req, nil
}

func (c *Client) handleRequest(req *http.Request) (*http.Response, error) {
	res, err := c.Client().Do(req)
	if err != nil {
		return nil, err
	}

	if err := responseError(res); err != nil {
		return nil, err
	}

	return res, nil
}

func unmarshalReader(r io.ReadCloser, out interface{}) error {
	defer r.Close()

	if out == nil {
		return nil
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}
