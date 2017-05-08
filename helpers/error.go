package helpers

import (
	"fmt"
	"os"
)

type AsyncErrorer func() error

func AsyncError(fn AsyncErrorer) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
	}
}
