package cli

import (
	"fmt"
	"strings"
	"unicode"
)

type Table struct {
	Headers []string
	Rows    [][]string
}

func (t *Table) AddHeader(headers ...string) {
	t.Headers = headers
}

func (t *Table) AddRow(row ...string) {
	t.Rows = append(t.Rows, row)
}

func (t *Table) Render() string {
	s := ""

	f := t.formatter()

	for _, rr := range t.rows() {
		ii := []interface{}{}

		for _, r := range rr {
			ii = append(ii, interface{}(r))
		}

		s += strings.TrimRightFunc(fmt.Sprintf(f, ii...), unicode.IsSpace) + "\n"
	}

	return s
}

func (t *Table) formatter() string {
	ll := t.lengths()

	f := []string{}

	for _, l := range ll {
		f = append(f, fmt.Sprintf("%%-%ds", l))
	}

	return strings.Join(f, "   ")
}

func (t *Table) lengths() []int {
	n := []int{}

	for _, r := range t.rows() {
		for i, c := range r {
			switch {
			case i >= len(n):
				n = append(n, len(c))
			case len(c) > n[i]:
				n[i] = len(c)
			}
		}
	}

	return n
}

func (t *Table) rows() [][]string {
	rr := [][]string{}

	if len(t.Headers) > 0 {
		rr = append(rr, t.Headers)
	}

	for _, r := range t.Rows {
		rr = append(rr, r)
	}

	return rr
}
