package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/zew/https-server/cfg"
)

var js = `
	<script src="/js/menu-and-form-control-keys.js?v=1637681051" ></script>
	<script src="/js/validation.js?v=1637681051"                 ></script>
`
var css = `
	<link href="/css/styles.css?v=1637681051"              rel="stylesheet" type="text/css" />
	<link href="/css/progress-bar-2.css?v=1637681051"      rel="stylesheet" type="text/css" />
	<link href="/css/styles-mobile.css?v=1637681051"       rel="stylesheet" type="text/css" />
	<link href="/css/styles-quest.css?v=1637681051"        rel="stylesheet" type="text/css" />
	<link href="/css/styles-quest-no-site-specified-a.css" rel="stylesheet" type="text/css" />
`

func init() {

	dirs := []string{"./static/js/", "./static/css/"}

	for idx, dir := range dirs {

		sb := &strings.Builder{}

		dirHandle, err := os.Open(dir) // open
		if err != nil {
			log.Fatalf("could not open %v, error %v", dir, err)
		}
		files, _ := dirHandle.Readdir(0) // get files
		if err != nil {
			log.Fatalf("could not read contents of %v, error %v", dir, err)
		}

		for _, file := range files {
			log.Printf("static %32v", file.Name())

			if idx == 0 {
				fmt.Fprintf(sb, `	<script src="/js/%v?v=%d" nonce="%d" ></script>`, file.Name(), cfg.Get().TS, cfg.Get().TS)
			}
			if idx == 1 {
				fmt.Fprintf(sb, `	<link href="/css/%s?v=%d" nonce="%d" rel="stylesheet" type="text/css" />`, file.Name(), cfg.Get().TS, cfg.Get().TS)
			}
			fmt.Fprint(sb, "\n")

		}

		if idx == 0 {
			js = sb.String()
		}
		if idx == 1 {
			css = sb.String()
		}

	}

}

func staticResources(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/js/") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(r.URL.Path, "/css/") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasPrefix(r.URL.Path, "/robots.txt") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	} else {
		return
	}
	pth := "./static" + r.URL.Path

	// bts, _ := ioutil.ReadFile(pth)
	// fmt.Fprint(w, bts)

	file, err := os.Open(pth)
	if err != nil {
		log.Printf("error opening %v, %v", pth, err)
		return
	}
	defer file.Close()

	n, err := io.Copy(w, file)
	if err != nil {
		log.Printf("error writing file into response: %v, %v", pth, err)
		return
	}

	log.Printf("%8v bytes written from %v", n, pth)

}
