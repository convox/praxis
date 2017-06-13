package helpers

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
)

var regexpCodeGrabber = regexp.MustCompile(`^26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd\[(\d+)\]\[(.*?)\]$`)

type codeGrabber struct {
	code *int
	err  *string
	w    io.Writer
}

func CodeGrabber(w io.Writer, code *int, err *string) io.Writer {
	return codeGrabber{code, err, w}
}

func CodeWrite(w io.Writer, code int) {
	fmt.Fprintf(w, "26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd[%d][]", code)
}

func CodeError(w io.Writer, code int, err error) {
	_, err = fmt.Fprintf(w, "26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd[%d][%s]", code, err.Error())
}

func (w codeGrabber) Write(data []byte) (int, error) {
	match := regexpCodeGrabber.FindSubmatch(data)

	if len(match) == 3 {
		i, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return 0, err
		}

		*w.code = i

		*w.err = ""
		if string(match[2]) != "" {
			*w.err = string(match[2])
		}

		if _, err := w.w.Write(regexpCodeGrabber.ReplaceAll(data, []byte{})); err != nil {
			return 0, err
		}

		return len(data), nil
	}

	return w.w.Write(data)
}
