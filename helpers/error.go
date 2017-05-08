package helpers

import (
	"fmt"
	"os"
)

func PrintError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
	}
}
