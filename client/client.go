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

	"golang.org/x/net/http2"
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
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if err := http2.ConfigureTransport(transport); err != nil {
		panic(err)
	}

	return &http.Client{
		Transport: transport,
	}
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

type GetOptions struct {
	Params   map[string]string
	Progress Progress
}

func (c *Client) Get(path string, out interface{}, opts GetOptions) error {
	req, err := c.Request("GET", path, nil)
	if err != nil {
		return err
	}

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

func (c *Client) GetReader(path string, opts GetOptions) (io.ReadCloser, error) {
	req, err := c.Request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Client().Do(req)
	if err != nil {
		return nil, err
	}

	if err := responseError(res); err != nil {
		return nil, err
	}

	return res.Body, nil
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

type DeleteOptions struct {
	Params   map[string]string
	Progress Progress
}

func (c *Client) Delete(path string, opts DeleteOptions) error {
	req, err := c.Request("DELETE", path, nil)
	if err != nil {
		return err
	}

	res, err := c.Client().Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if err := responseError(res); err != nil {
		return err
	}

	return nil
}
