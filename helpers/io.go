package helpers

import (
	"io"
	"net/http"
)

type ReadWriter struct {
	io.Reader
	io.Writer
}

func HalfPipe(w io.Writer, r io.Reader) error {
	defer func() {
		if c, ok := w.(io.Closer); ok {
			c.Close()
		}
	}()

	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, err := w.Write(buf[0:n]); err != nil {
				if err != io.ErrClosedPipe {
					return err
				}
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

func HalfPipeAsync(w io.Writer, r io.Reader, ch chan error) {
	ch <- HalfPipe(w, r)
}

func Pipe(a, b io.ReadWriter) error {
	ch := make(chan error)

	go HalfPipeAsync(a, b, ch)
	go HalfPipeAsync(b, a, ch)

	if err := <-ch; err != nil {
		return err
	}

	if err := <-ch; err != nil {
		return err
	}

	return nil
}
