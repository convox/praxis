package helpers

import (
	"io"
	"net/http"
)

func Pipe(a, b io.ReadWriter) error {
	ch := make(chan error)

	go StreamAsync(a, b, ch)
	go StreamAsync(b, a, ch)

	if err := <-ch; err != nil {
		return err
	}

	if err := <-ch; err != nil {
		return err
	}

	return nil
}

func Stream(w io.Writer, r io.Reader) error {
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, err := w.Write(buf[0:n]); err != nil {
				return err
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func StreamAsync(w io.Writer, r io.Reader, ch chan error) {
	err := Stream(w, r)

	if c, ok := w.(io.Closer); ok {
		c.Close()
	}

	ch <- err
}
