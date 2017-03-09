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

	return s.Reader.Read(data)
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
	if err != nil {
		return n, err
	}

	if f, ok := s.Writer.(http.Flusher); ok {
		f.Flush()
	}

	return n, err
}
