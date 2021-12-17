package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/zew/https-server/cfg"
	"github.com/zew/https-server/gzipper"
)

func init() {

	log.SetFlags(log.Lshortfile | log.Llongfile)

	dirs := []string{"./static/js/", "./static/css/"}

	exceptions := map[string]bool{
		// "service-worker.js": true,
	}

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

			if exceptions[file.Name()] {
				log.Printf("\t  static skipping exception %v", file.Name())
				continue
			}

			if strings.HasSuffix(file.Name(), ".gzip") {
				// log.Printf("\t  static skipping %v", file.Name())
				continue
			}

			log.Printf("\tstatic %v", file.Name())

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
					log.Printf("could not create gzWriter for %v, %v", fn, err)
				}
				defer gzw.Close()
				// gzw.WriteString("gzipper.go created this file.\n")
				gzw.WriteFile(fn)
			}
			fnc()

		}

		if idx == 0 {
			cfg.Get().JS = sb.String()
		}
		if idx == 1 {
			cfg.Get().CSS = sb.String()
		}

	}

}

func staticResources(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/js/") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(r.URL.Path, "/service-worker.js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(r.URL.Path, "/css/") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasPrefix(r.URL.Path, "/img/") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.HasPrefix(r.URL.Path, "/json/") {
		w.Header().Set("Content-Type", "application/json")
	} else if strings.HasPrefix(r.URL.Path, "/robots.txt") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	} else {
		return
	}

	// andrewlock.net/adding-cache-control-headers-to-static-files-in-asp-net.core/
	w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))

	pth := "./static" + r.URL.Path

	// bts, _ := ioutil.ReadFile(pth)
	// fmt.Fprint(w, bts)

	if cfg.Get().PrecompressedGZIP {
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

	n, err := io.Copy(w, file)
	if err != nil {
		log.Printf("error writing file into response: %v, %v", pth, err)
		return
	}

	log.Printf("%8v bytes written from %v", n, pth)

}
