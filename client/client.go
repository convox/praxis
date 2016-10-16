package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Client struct {
	Endpoint string
}

func New(endpoint string) *Client {
	return &Client{
		Endpoint: endpoint,
	}
}

func (c *Client) Client() *http.Client {
	client := &http.Client{}

	var config *tls.Config

	// FIXME: better verification
	config = &tls.Config{
		InsecureSkipVerify: true,
	}

	client.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: config,
	}

	return client
}

func (c *Client) Request(method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, fmt.Sprintf("https://%s%s%s", u.Host, u.Path, path), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	req.SetBasicAuth("convox", "")

	return req, nil
}

type PostOptions struct {
	Files    map[string]io.Reader
	Params   map[string]string
	Progress Progress
}

func (c *Client) Post(path string, out interface{}, opts PostOptions) error {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	for name, file := range opts.Files {
		part, err := writer.CreateFormFile(name, "binary-data")
		if err != nil {
			return err
		}

		if _, err = io.Copy(part, file); err != nil {
			return err
		}
	}

	for name, value := range opts.Params {
		writer.WriteField(name, value)
	}

	if err := writer.Close(); err != nil {
		return err
	}

	br := io.Reader(body)

	if opts.Progress != nil {
		opts.Progress.Start(int64(body.Len()))

		defer opts.Progress.Finish()

		br = NewProgressReader(br, opts.Progress.Progress)
	}

	req, err := c.Request("POST", path, br)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := c.Client().Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if err := responseError(res); err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	if out != nil {
		err = json.Unmarshal(data, out)

		if err != nil {
			return err
		}
	}

	return nil
}
