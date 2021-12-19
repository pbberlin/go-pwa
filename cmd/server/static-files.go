package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/zew/https-server/pkg/cfg"
	"github.com/zew/https-server/pkg/gzipper"
)

func init() {
	// cfg.RunHandleFunc(prepareStatic, "/prepare-static")
}

func prepareStatic(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "prepare static files\n")

	if cfg.Get().TS == 0 {
		panic("init - prepareStatic called before cfg::init().Load")
	}

	dirs := []string{
		"./static/js/",
		"./static/css/",
		"./static/", // the service-worker.js
	}

	exceptions := map[string]bool{
		// "service-worker.js": true,
	}

	for idx, dir := range dirs {

		fmt.Fprintf(w, "start dir %v\n", dir)

		sb := &strings.Builder{}

		dirHandle, err := os.Open(dir) // open
		if err != nil {
			fmt.Fprintf(w, "could not open %v, error %v\n", dir, err)
			return
		}
		files, _ := dirHandle.Readdir(0) // get files
		if err != nil {
			fmt.Fprintf(w, "could not read contents of %v, error %v\n", dir, err)
			return
		}

		for _, file := range files {

			if exceptions[file.Name()] {
				fmt.Fprintf(w, "\t  skipping exception %v\n", file.Name())
				continue
			}

			if strings.HasSuffix(file.Name(), ".gzip") {
				// fmt.Fprintf(w, "\t  static skipping %v\n", file.Name())
				continue
			}

			if file.IsDir() {
				continue
			}

			fmt.Fprintf(w, "\t %v\n", file.Name())

			if idx == 0 {
				fmt.Fprintf(sb, `	<script src="/js/%v?v=%d" nonce="%d" ></script>`, file.Name(), cfg.Get().TS, cfg.Get().TS)
			}
			if idx == 1 {
				fmt.Fprintf(sb, `	<link href="/css/%s?v=%d" nonce="%d" rel="stylesheet" type="text/css" />`, file.Name(), cfg.Get().TS, cfg.Get().TS)
			}
			fmt.Fprint(sb, "\n")

			//
			// closure because of defer
			fnc := func() {
				fn := path.Join(dir, file.Name())
				gzw, err := gzipper.New(fn)
				if err != nil {
					fmt.Fprintf(w, "could not create gzWriter for %v, %v\n", fn, err)
					return
				}
				defer gzw.Close()
				// gzw.WriteString("gzipper.go created this file.\n")
				gzw.WriteFile(fn)
			}
			fnc()

		}

		if idx == 0 {
			cfg.Set().JS = sb.String()
		}
		if idx == 1 {
			cfg.Set().CSS = sb.String()
		}

		fmt.Fprintf(w, "stop dir %v\n\n", dir)

	}

}

func staticResources(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/js/") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	} else if strings.HasPrefix(r.URL.Path, "/service-worker.js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(r.URL.Path, "/css/") {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	} else if strings.HasPrefix(r.URL.Path, "/img/") {
		if strings.HasPrefix(r.URL.Path, "/img/favicon.ico") {
			w.Header().Set("Content-Type", "image/x-icon")
		} else {
			w.Header().Set("Content-Type", "image/png")
		}
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	} else if strings.HasPrefix(r.URL.Path, "/json/") {
		w.Header().Set("Content-Type", "application/json")
	} else if strings.HasPrefix(r.URL.Path, "/robots.txt") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	} else {
		return
	}

	pth := "./static" + r.URL.Path

	// bts, _ := os.ReadFile(pth)
	// fmt.Fprint(w, bts)

	if cfg.Get().PrecompressGZIP {
		if strings.HasSuffix(pth, ".js") || strings.HasSuffix(pth, ".css") {
			// not for images or json
			pth += ".gzip"
			w.Header().Set("Content-Encoding", "gzip")
		}
	}

	file, err := os.Open(pth)
	if err != nil {
		log.Printf("error opening %v, %v", pth, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("error writing file into response: %v, %v", pth, err)
		return
	}
	// log.Printf("%8v bytes written from %v", n, pth)

}
