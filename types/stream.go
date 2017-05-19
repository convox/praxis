package types

import (
	"io"
	"net/http"
)

type Stream struct {
	Reader io.Reader
	Writer io.Writer
}

func (s Stream) Read(data []byte) (int, error) {
	if s.Reader == nil {
		return 0, io.EOF
	}

	n, err := s.Reader.Read(data)

	return n, err
}

func (s Stream) Write(data []byte) (int, error) {
	if s.Writer == nil {
		return 0, io.EOF
	}

	// for i, b := range data {
	//   if b == 13 {
	//     data[i] = 10
	//   }
	// }

	n, err := s.Writer.Write(data)
	if err == io.ErrClosedPipe {
		return n, nil
	}
	if err != nil {
		return n, err
	}

	if n > 0 {
		if f, ok := s.Writer.(http.Flusher); ok {
			f.Flush()
		}
	}

	return n, err
}

func (s Stream) Close() error {
	if c, ok := s.Writer.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}

	if c, ok := s.Reader.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}
