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

	dirs := []string{
		"./app-bucket/js/",
		"./app-bucket/css/",
		"./app-bucket/", // the service-worker.js
	}

	exceptions := map[string]bool{
		// "service-worker.js": true,
	}

	for idx, dirSrc := range dirs {

		dirDst := path.Join(dirSrc, fmt.Sprint(cfg.Get().TS))
		err := os.MkdirAll(dirDst, 0755)
		if err != nil {
			fmt.Fprintf(w, "could not create dir %v, %v\n", dirSrc, err)
			return
		}

		fmt.Fprintf(w, "start src dir %v\n", dirSrc)

		sb := &strings.Builder{}

		dirHandle, err := os.Open(dirSrc) // open
		if err != nil {
			fmt.Fprintf(w, "could not open %v, error %v\n", dirSrc, err)
			return
		}
		files, _ := dirHandle.Readdir(0) // get files
		if err != nil {
			fmt.Fprintf(w, "could not read contents of %v, error %v\n", dirSrc, err)
			return
		}

		for _, file := range files {

			if exceptions[file.Name()] {
				fmt.Fprintf(w, "\t  skipping exception %v\n", file.Name())
				continue
			}

			if strings.HasSuffix(file.Name(), ".gzip") {
				continue
			}

			if file.IsDir() {
				continue
			}

			fmt.Fprintf(w, "\t %v\n", file.Name())

			if idx == 0 {
				// fmt.Fprintf(sb, `	<script src="/js/%v?v=%d" nonce="%d" ></script>`, file.Name(), cfg.Get().TS, cfg.Get().TS)
				fmt.Fprintf(sb, `	<script src="/js/%s/%v" nonce="%s" ></script>`, cfg.Get().TS, file.Name(), cfg.Get().TS)
			}
			if idx == 1 {
				// fmt.Fprintf(sb, `	<link href="/css/%s?v=%d" nonce="%d" rel="stylesheet" type="text/css" />`, file.Name(), cfg.Get().TS, cfg.Get().TS)
				fmt.Fprintf(sb, `	<link href="/css/%s/%s" nonce="%s" rel="stylesheet" type="text/css" />`, cfg.Get().TS, file.Name(), cfg.Get().TS)
			}
			fmt.Fprint(sb, "\n")

			//
			// closure because of defer
			fnc := func() {
				fnSrc := path.Join(dirSrc, file.Name())
				fnDst := path.Join(dirDst, file.Name())
				gzw, err := gzipper.New(fnDst)
				if err != nil {
					fmt.Fprint(w, err)
					return
				}
				defer gzw.Close()
				// gzw.WriteString("gzipper.go created this file.\n")
				err = gzw.WriteFile(fnSrc)
				if err != nil {
					fmt.Fprint(w, err)
					return
				}
			}
			fnc()

		}

		if idx == 0 {
			cfg.Set().JS = sb.String()
		}
		if idx == 1 {
			cfg.Set().CSS = sb.String()
		}

		fmt.Fprintf(w, "stop src dir  %v\n\n", dirSrc)

	}

}

func serveStatic(w http.ResponseWriter, r *http.Request) {

	switch {

	//
	// special files
	case strings.HasPrefix(r.URL.Path, "/robots.txt"):
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	case strings.HasPrefix(r.URL.Path, "/img/favicon.ico"):
		w.Header().Set("Content-Type", "image/x-icon")
	case strings.HasPrefix(r.URL.Path, "/service-worker.js"):
		w.Header().Set("Content-Type", "application/javascript")

	//
	// types .js, .css, .webp, .json
	case strings.HasPrefix(r.URL.Path, "/js/"):
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	case strings.HasPrefix(r.URL.Path, "/css/"):
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	case strings.HasPrefix(r.URL.Path, "/img/"):
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))
	case strings.HasPrefix(r.URL.Path, "/json/"):
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "unsupported static directory")
		return
	}

	pth := "./app-bucket" + r.URL.Path

	if cfg.Get().PrecompressGZIP {
		// not for images or json
		if strings.HasPrefix(r.URL.Path, "/js/") || strings.HasPrefix(r.URL.Path, "/css/") {
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
		log.Printf("error writing filecontents into response: %v, %v", pth, err)
		return
	}
	// log.Printf("%8v bytes written from %v", n, pth)

}
