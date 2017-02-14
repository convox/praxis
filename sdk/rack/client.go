package rack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Client struct {
	Host    string
	Key     string
	Rack    string
	Socket  string
	Version string
}

func (c *Client) GetStream(path string) (io.ReadCloser, error) {
	req, err := c.Request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Client().Do(req)
	if err != nil {
		return nil, err
	}

	if !res.ProtoAtLeast(2, 0) {
		return nil, fmt.Errorf("server did not respond with http/2")
	}

	if err := responseError(res); err != nil {
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

func (c *Client) Client() *http.Client {
	client := http.DefaultClient

	if c.Socket != "" {
		client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, proto, addr string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext(ctx, "unix", c.Socket)
			},
		}
	}

	return client
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

func unmarshalReader(r io.Reader, out interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}
