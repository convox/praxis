package rack

func New() *Client {
	return &Client{
		Host: "localhost:9666",
	}
}
