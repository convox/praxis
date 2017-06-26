package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"html/template"
	"net"
	"path"
	"strings"
)

func formationHelpers() template.FuncMap {
	return template.FuncMap{
		"apex": func(domain string) string {
			parts := strings.Split(domain, ".")
			for i := 0; i < len(parts)-1; i++ {
				d := strings.Join(parts[i:], ".")
				if mx, err := net.LookupMX(d); err == nil && len(mx) > 0 {
					return d
				}
			}
			return domain
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"priority": func(app, service string) uint32 {
			return crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s-%s", app, service))) % 50000
		},
		"resource": func(s string) string {
			return upperName(s)
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"volumeFrom": func(s string) string {
			parts := strings.SplitN(s, ":", 2)
			switch len(parts) {
			case 1:
				return path.Join("/volumes", s)
			case 2:
				return parts[0]
			}
			return fmt.Sprintf("invalid volume %q", s)
		},
		"volumeTo": func(s string) string {
			parts := strings.SplitN(s, ":", 2)
			switch len(parts) {
			case 1:
				return s
			case 2:
				return parts[1]
			}
			return fmt.Sprintf("invalid volume %q", s)
		},
	}
}

func formationTemplate(name string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer

	tn := fmt.Sprintf("%s.json.tmpl", name)
	tf := fmt.Sprintf("provider/aws/formation/%s", tn)

	t, err := template.New(tn).Funcs(formationHelpers()).ParseFiles(tf)
	if err != nil {
		return nil, err
	}

	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	var v interface{}

	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		switch t := err.(type) {
		case *json.SyntaxError:
			return nil, jsonSyntaxError(t, buf.Bytes())
		}
		return nil, err
	}

	return json.MarshalIndent(v, "", "  ")
}

func jsonSyntaxError(err *json.SyntaxError, data []byte) error {
	start := bytes.LastIndex(data[:err.Offset], []byte("\n")) + 1
	line := bytes.Count(data[:start], []byte("\n"))
	pos := int(err.Offset) - start - 1
	ltext := strings.Split(string(data), "\n")[line]

	return fmt.Errorf("json syntax error: line %d pos %d: %s: %s", line, pos, err.Error(), ltext)
}
