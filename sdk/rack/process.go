package rack

import "io"

type Process struct {
	Id string
}

type ProcessRunOptions struct {
	Command string
	Input   io.Reader
	Output  io.Writer
}

func (c *Client) ProcessRun(app, service string, opts ProcessRunOptions) (*Process, error) {
	return &Process{Id: "1234"}, nil
}
