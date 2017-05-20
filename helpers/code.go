package helpers

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
)

var regexpCodeGrabber = regexp.MustCompile(`^26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd\[(\d+)\]$`)

type CodeGrabber struct {
	Code   *int
	Writer io.Writer
}

func CodeWrite(w io.Writer, code int) {
	fmt.Fprintf(w, "26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd[%d]", code)
}

func (w CodeGrabber) Write(data []byte) (int, error) {
	match := regexpCodeGrabber.FindSubmatch(data)

	if len(match) == 2 {
		i, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return 0, err
		}

		*w.Code = i

		if _, err := w.Writer.Write(regexpCodeGrabber.ReplaceAll(data, []byte{})); err != nil {
			return 0, err
		}

		return len(data), nil
	}

	return w.Writer.Write(data)
}
