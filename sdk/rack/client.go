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
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type Client struct {
	Endpoint *url.URL
	Key      string
	Socket   string
	Version  string
}

type Headers map[string]string
type Params map[string]interface{}
type Query map[string]string

type RequestOptions struct {
	Body    io.Reader
	Headers Headers
	Params  Params
	Query   Query
}

func (o *RequestOptions) Querystring() string {
	uv := url.Values{}

	for k, v := range o.Query {
		uv.Add(k, v)
	}

	return uv.Encode()
}

func (o *RequestOptions) Reader() (io.Reader, error) {
	if o.Body != nil && len(o.Params) > 0 {
		return nil, fmt.Errorf("cannot specify both Body and Params")
	}

	if o.Body != nil {
		return o.Body, nil
	}

	u := url.Values{}

	for k, v := range o.Params {
		switch t := v.(type) {
		case string:
			u.Set(k, t)
		case []string:
			for _, s := range t {
				u.Add(k, s)
			}
		case time.Time:
			u.Set(k, t.Format(sortableTime))
		default:
			return nil, fmt.Errorf("unknown param type: %T", t)
		}
	}

	return bytes.NewReader([]byte(u.Encode())), nil
}

func (o *RequestOptions) ContentType() string {
	if o.Body == nil {
		return "application/x-www-form-urlencoded"
	}

	return "application/octet-stream"
}

func (c *Client) GetStream(path string, opts RequestOptions) (*http.Response, error) {
	req, err := c.Request("GET", path, opts)
	if err != nil {
		return nil, err
	}

	return c.handleRequest(req)
}

func (c *Client) Get(path string, opts RequestOptions, out interface{}) error {
	res, err := c.GetStream(path, opts)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return unmarshalReader(res.Body, out)
}

func (c *Client) PostStream(path string, opts RequestOptions) (*http.Response, error) {
	req, err := c.Request("POST", path, opts)
	if err != nil {
		return nil, err
	}

	res, err := c.handleRequest(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) Post(path string, opts RequestOptions, out interface{}) error {
	res, err := c.PostStream(path, opts)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return unmarshalReader(res.Body, out)
}

func (c *Client) PutStream(path string, opts RequestOptions) (*http.Response, error) {
	req, err := c.Request("PUT", path, opts)
	if err != nil {
		return nil, err
	}

	return c.handleRequest(req)
}

func (c *Client) Put(path string, opts RequestOptions, out interface{}) error {
	res, err := c.PutStream(path, opts)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return unmarshalReader(res.Body, out)
}

func (c *Client) Delete(path string, opts RequestOptions, out interface{}) error {
	req, err := c.Request("DELETE", path, opts)
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
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 2 * time.Second,
	}

	t := &http.Transport{
		DialContext: func(ctx context.Context, proto, addr string) (net.Conn, error) {
			if c.Socket != "" {
				proto = "unix"
				addr = c.Socket
			}
			return dialer.DialContext(ctx, proto, addr)
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

func (c *Client) Request(method, path string, opts RequestOptions) (*http.Request, error) {
	qs := opts.Querystring()

	r, err := opts.Reader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s?%s", c.Endpoint.String(), path, qs), r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", opts.ContentType())
	req.Header.Add("Version", c.Version)

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	req.SetBasicAuth(c.Endpoint.User.Username(), "")

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

func responseError(res *http.Response) error {
	if !res.ProtoAtLeast(2, 0) {
		return fmt.Errorf("server did not respond with http/2")
	}

	if res.StatusCode < 400 {
		return nil
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	msg := strings.TrimSpace(string(data))

	if len(msg) > 0 {
		return fmt.Errorf(msg)
	}

	return fmt.Errorf("response status %d", res.StatusCode)
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
