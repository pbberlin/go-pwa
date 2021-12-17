package main

import (
	"fmt"
	"net/http"
	"strings"
)

func home(w http.ResponseWriter, r *http.Request) {
	cnt := &strings.Builder{}
	fmt.Fprintf(cnt, "<p>Hello, TLS user </p>\n")
	fmt.Fprintf(cnt, "<p>Your config: </p>\n")
	fmt.Fprintf(cnt, "<pre id='tls-config'>%+v  </pre>\n", r.TLS)

	sc := &Scaffold{}
	sc.Title = "Home page"
	sc.render(w, cnt)
}

func plain(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "This is an example server.\n")
}
