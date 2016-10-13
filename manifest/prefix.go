package manifest

import (
	"fmt"
	"io"
	"strings"
)

func (m *Manifest) system(format string, args ...interface{}) {
	m.prefix("system").Write([]byte(fmt.Sprintf(format+"\n", args...)))
}

func (m *Manifest) prefix(name string) io.Writer {
	longest := 6

	for _, s := range m.Services {
		if len(s.Name) > longest {
			longest = len(s.Name)
		}
	}

	return prefixer{
		Prefix:     fmt.Sprintf(fmt.Sprintf("%%-%ds | ", longest), name),
		Writer:     Writer,
		prefixNext: true,
	}
}

type prefixer struct {
	Prefix string
	Writer io.Writer

	prefixNext bool
}

func (p prefixer) Write(data []byte) (int, error) {
	r := ""

	if p.prefixNext {
		r += p.Prefix
		p.prefixNext = false
	}

	s := string(data)

	if data[len(data)-1] == '\n' {
		p.prefixNext = true
		s = s[0 : len(s)-1]
	}

	r += strings.Replace(s, "\n", fmt.Sprintf("\n%s", p.Prefix), -1)

	if data[len(data)-1] == '\n' {
		r += "\n"
	}

	if _, err := p.Writer.Write([]byte(r)); err != nil {
		return 0, err
	}

	return len(data), nil
}
