package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

type PrefixWriter struct {
	Buffer bytes.Buffer
	Writer func(string) error
}

var writeLock sync.Mutex

func (m *Manifest) WriteLine(line string) {
	writeLock.Lock()
	defer writeLock.Unlock()
	fmt.Println(line)
}

func (m *Manifest) Writef(label string, format string, args ...interface{}) {
	m.Writer(label, os.Stdout).Write([]byte(fmt.Sprintf(format, args...)))
}

var lock sync.Mutex

func (m *Manifest) Writer(label string, w io.Writer) PrefixWriter {
	prefix := []byte(fmt.Sprintf(fmt.Sprintf("%%-%ds | ", m.prefixLength()), label))

	return PrefixWriter{
		Writer: func(s string) error {
			lock.Lock()
			defer lock.Unlock()

			if _, err := w.Write(prefix); err != nil {
				return err
			}

			if _, err := w.Write([]byte(s)); err != nil {
				return err
			}

			return nil
		},
	}
}

func (w PrefixWriter) Write(p []byte) (int, error) {
	q := bytes.Replace(p, []byte{10, 13}, []byte{10}, -1)

	if _, err := w.Buffer.Write(q); err != nil {
		return 0, err
	}

	if idx := bytes.Index(w.Buffer.Bytes(), []byte{10}); idx > -1 {
		if err := w.Writer(string(w.Buffer.Next(idx + 1))); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (w PrefixWriter) Writef(format string, args ...interface{}) error {
	_, err := w.Write([]byte(fmt.Sprintf(format, args...)))
	return err
}

func (m *Manifest) prefixLength() int {
	max := 6 // convox

	for _, s := range m.Services {
		if len(s.Name) > max {
			max = len(s.Name)
		}
	}

	return max
}
