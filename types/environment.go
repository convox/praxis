package types

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Environment map[string]string

func (e *Environment) Pairs(pairs []string) error {
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)

		if len(parts) != 2 {
			return fmt.Errorf("invalid environment: %s", p)
		}

		(*e)[parts[0]] = parts[1]
	}

	return nil
}

func (e *Environment) Read(r io.Reader) error {
	pairs := []string{}

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if pair := strings.TrimSpace(scanner.Text()); pair != "" {
			pairs = append(pairs, pair)
		}
	}

	return e.Pairs(pairs)
}
