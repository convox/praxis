package controllers

import (
	"context"

	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

const (
	sortableTime = "20060102.150405.000000000"
)

var (
	Provider types.Provider
)

func Setup() error {
	p, err := provider.FromEnv()
	if err != nil {
		return errors.WithStack(err)
	}

	Provider = p

	go Provider.WithContext(context.WithValue(context.Background(), "request.id", "workers")).Workers()

	return nil
}

// func stream(w io.Writer, r io.Reader) error {
//   buf := make([]byte, 1024)

//   for {
//     n, err := r.Read(buf)
//     if n > 0 {
//       if _, err := w.Write(buf[0:n]); err != nil {
//         return err
//       }
//       if f, ok := w.(http.Flusher); ok {
//         f.Flush()
//       }
//     }
//     if err == io.EOF {
//       return nil
//     }
//     if err != nil {
//       return err
//     }
//   }
// }
