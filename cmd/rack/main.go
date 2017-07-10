package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/server"
)

func main() {
	s := server.New()

	if err := s.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	// go http.ListenAndServe(":3001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//   u := r.URL

	//   u.Host = strings.Split(r.Host, ":")[0]
	//   u.Scheme = "https"

	//   http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	// }))

	if err := s.Listen("tcp", ":3000"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
