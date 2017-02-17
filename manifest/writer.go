package manifest

import (
	"fmt"
	"io"
	"os"

	"github.com/kr/text"
)

type PrefixWriter struct {
	Label  string
	Writer io.Writer
	prefix string
}

func (m *Manifest) PrefixWriter(w io.Writer, label string) PrefixWriter {
	prefix := fmt.Sprintf(fmt.Sprintf("%%-%ds | ", m.prefixLength()), label)

	return PrefixWriter{
		Label:  label,
		Writer: text.NewIndentWriter(w, []byte(prefix)),
		prefix: prefix,
	}
}

func (m *Manifest) Write(p []byte) (int, error) {
	prefix := fmt.Sprintf(fmt.Sprintf("%%-%ds | ", m.prefixLength()), "convox")

	if _, err := os.Stdout.Write([]byte(prefix)); err != nil {
		return 0, err
	}

	return os.Stdout.Write(p)
}

func (m *Manifest) Writef(format string, args ...interface{}) error {
	return m.WriteString(fmt.Sprintf(format, args...))
}

func (m *Manifest) WriteString(s string) error {
	_, err := m.Write([]byte(s))
	return err
}

func (w PrefixWriter) Write(p []byte) (int, error) {
	q := []byte{}

	// inject prefix after line feeds unless they are followed by
	// a carriage return
	for i, b := range p {
		q = append(q, b)
		if b == 13 && i < len(p)-1 && p[i+1] != 10 {
			q = append(q, []byte(w.prefix)...)
		}
	}

	if _, err := w.Writer.Write(q); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w PrefixWriter) Writef(format string, args ...interface{}) error {
	return w.WriteString(fmt.Sprintf(format, args...))
}

func (w PrefixWriter) WriteString(s string) error {
	_, err := w.Write([]byte(s))
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
