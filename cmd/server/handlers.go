package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func home(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("parsing form %v", err)
	}
	if r.Form.Get("refuse") != "" {
		w.WriteHeader(401)
		log.Print("offline 401")
		return
	}

	cnt := &strings.Builder{}
	fmt.Fprintf(cnt, "<p>Hello, TLS user </p>\n")
	fmt.Fprintf(cnt, "<p>Your config: </p>\n")
	fmt.Fprintf(cnt, "<pre id='tls-config'>%+v  </pre>\n", r.TLS)

	sc := &Scaffold{}
	sc.Title = "Home page"
	sc.render(w, cnt)
}

func offline(w http.ResponseWriter, r *http.Request) {
	cnt := &strings.Builder{}
	fmt.Fprintf(cnt,
		`<p>
			This beautiful offline page<br>
			provides a way better "user experience".

		</p>`)

	sc := &Scaffold{}
	sc.Title = "You are offline"
	sc.render(w, cnt)
}

func plain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "This is an example server.\n")
}

func saveJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.PostFormValue("refuse") != "" {
		return
	}
	fmt.Fprintf(w, "%+v", r.PostForm)
}
