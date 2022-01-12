package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
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

var staticFilesSeed = []string{
	"/index.html",
	"/offline.html",

	// ... dynamic css, js, webp

	// example for external res
	// "https://fonts.google.com/icon?family=Material+Icons",

}

func filesOfDir(dirSrc string) ([]fs.FileInfo, error) {
	dirHandle, err := os.Open(dirSrc) // open
	if err != nil {
		return nil, fmt.Errorf("could not open dir %v, error %v\n", dirSrc, err)
	}
	files, err := dirHandle.Readdir(0) // get files
	if err != nil {
		return nil, fmt.Errorf("could not read dir contents of %v, error %v\n", dirSrc, err)
	}
	return files, nil
}

// prepareServiceWorker
func prepareServiceWorker(w http.ResponseWriter, req *http.Request) {

	srcPth := "./app-bucket/tpl/service-worker.tpl.js"
	t, err := template.ParseFiles(srcPth)
	if err != nil {
		fmt.Fprintf(w, "could not parse service worker template %v\n", err)
		return
	}

	staticFiles := make([]string, len(staticFilesSeed), 16)
	copy(staticFiles, staticFilesSeed)

	dirs := []string{
		"./app-bucket/js/",
		"./app-bucket/css/",
		"./app-bucket/img/",
	}

	for _, dirSrc := range dirs {

		files, err := filesOfDir(dirSrc)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		for _, file := range files {

			if file.IsDir() {
				continue
			}

			if strings.HasSuffix(file.Name(), ".png") {
				continue
			}
			if strings.HasPrefix(file.Name(), "tmp-") {
				continue
			}

			// fmt.Fprintf(w, "\t %v\n", file.Name())
			pth := path.Join(dirSrc, cfg.Get().TS, file.Name())
			pth = strings.Replace(pth, "app-bucket/", "./", 1)
			staticFiles = append(staticFiles, pth)
		}
	}

	bts := &bytes.Buffer{}

	data := struct {
		Version     string
		ListOfFiles template.HTML // neither .JS nor .JSStr do work
	}{
		Version:     cfg.Get().TS,
		ListOfFiles: template.HTML("'" + strings.Join(staticFiles, "',\n  '") + "'"),
	}

	err = t.Execute(bts, data)
	if err != nil {
		fmt.Fprintf(w, "could not execute service worker template %v\n", err)
		return
	}

	dstPth := "./app-bucket/js-service-worker/service-worker.js"
	os.WriteFile(dstPth, bts.Bytes(), 0644)
}

//
//
func prepareStatic(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "prepare static files\n")

	prepareServiceWorker(w, req)

	//
	//
	dirs := []string{
		"./app-bucket/js/",
		"./app-bucket/js-service-worker/",
		"./app-bucket/css/",
		"./app-bucket/img/",
	}

	for dirIdx, dirSrc := range dirs {

		dirDst := path.Join(dirSrc, fmt.Sprint(cfg.Get().TS))
		err := os.MkdirAll(dirDst, 0755)
		if err != nil {
			fmt.Fprintf(w, "could not create dir %v, %v\n", dirSrc, err)
			return
		}

		fmt.Fprintf(w, "start src dir %v\n", dirSrc)

		sb := &strings.Builder{}

		files, err := filesOfDir(dirSrc)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		for _, file := range files {

			if strings.HasSuffix(file.Name(), ".gzip") {
				continue
			}

			if file.IsDir() {
				continue
			}

			if dirIdx == 0 {
				fmt.Fprintf(sb, `	<script src="/js/%s/%v" nonce="%s" ></script>`, cfg.Get().TS, file.Name(), cfg.Get().TS)
				fmt.Fprint(sb, "\n")
			}
			if dirIdx == 2 {
				fmt.Fprintf(sb, `	<link href="/css/%s/%s" nonce="%s" rel="stylesheet" type="text/css" />`, cfg.Get().TS, file.Name(), cfg.Get().TS)
				fmt.Fprint(sb, "\n")
			}

			fnSrc := path.Join(dirSrc, file.Name())
			fnDst := path.Join(dirDst, file.Name())

			//
			if dirIdx < 3 {
				//
				// closure because of defer
				zipIt := func() {
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
				zipIt()
			}
			//
			if dirIdx == 3 {
				ext := path.Ext(file.Name())
				if ext != ".webp" && ext != ".ico" {
					continue
				}

				bts, err := os.ReadFile(fnSrc)
				if err != nil {
					fmt.Fprint(w, err)
					return
				}
				os.WriteFile(fnDst, bts, 0644)
			}

			fmt.Fprintf(w, "\t %v\n", file.Name())

		}

		if dirIdx == 0 {
			cfg.Set().JS = sb.String()
		}
		if dirIdx == 2 {
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
		r.URL.Path = fmt.Sprintf("/js-service-worker/%v/service-worker.js", cfg.Get().TS)

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

		// todo: versioning of images
		// replacement := fmt.Sprintf("/js/%v/", cfg.Get().TS)
		// r.URL.Path = strings.Replace(r.URL.Path, "/js/", replacement, 1)

	case strings.HasPrefix(r.URL.Path, "/json/"):
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "unsupported static directory")
		return
	}

	pth := "./app-bucket" + r.URL.Path

	if cfg.Get().PrecompressGZIP {
		// not for images, not for json
		if strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/js-service-worker/") ||
			strings.HasPrefix(r.URL.Path, "/css/") {
			pth += ".gzip"
			w.Header().Set("Content-Encoding", "gzip")
		}
	}

	file, err := os.Open(pth)
	if err != nil {
		log.Printf("error opening %v, %v", pth, err)
		log.Printf("\tfrom %+v", r.Header.Get("Referer"))
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
