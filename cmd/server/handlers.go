package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/metrics"
	"sort"
	"strings"

	"github.com/pbberlin/go-pwa/pkg/db"
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

func DBTestData(w http.ResponseWriter, r *http.Request) {
	db.TestData()
}

func saveJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.PostFormValue("refuse") != "" {
		return
	}
	fmt.Fprintf(w, "%+v", r.PostForm)
}

func layoutTemplateForJS(w http.ResponseWriter, r *http.Request) {
	cnt := &strings.Builder{}
	fmt.Fprintf(cnt, "${content}")
	sc := &Scaffold{}
	sc.Title = "${headerTitle}"
	sc.Desc = "${headerDesc}"
	sc.render(w, cnt)
}

var mtrcs = map[string]string{
	"gcPauses":       "/gc/pauses:seconds",
	"schedLatencies": "/sched/latencies:seconds",
	"memoryHeapFree": "/memory/classes/heap/free:bytes",
}

// golangMetrics uses new 2021 package metrics
func golangMetrics(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	srt := make([]string, 0, len(mtrcs))
	for key := range mtrcs {
		srt = append(srt, key)
	}
	sort.Strings(srt)

	for _, key := range srt {
		val := mtrcs[key]

		// create smpl and read it
		smpl := make([]metrics.Sample, 1)
		smpl[0].Name = val
		metrics.Read(smpl)

		if smpl[0].Value.Kind() == metrics.KindBad {
			fmt.Fprintf(w, "\tmetric %q unsupported\n", val)
		}

		// uintVal := smpl[0].Value.Uint64()
		fmt.Fprintf(w, "%-24v as %T: %d\n", key, smpl[0].Value, smpl[0].Value)
	}
}
