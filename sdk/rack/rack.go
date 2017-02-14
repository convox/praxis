package rack

func New() *Client {
	return &Client{
		Host: "http2.golang.org",
	}
}
