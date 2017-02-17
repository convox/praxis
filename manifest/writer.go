package manifest

import (
	"fmt"
	"io"

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

func (m *Manifest) prefixLength() int {
	max := 6 // convox

	for _, s := range m.Services {
		if len(s.Name) > max {
			max = len(s.Name)
		}
	}

	return max
}
