package static

import (
	"bytes"
	"fmt"
	htmltpl "html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	texttpl "text/template"

	"github.com/zew/https-server/pkg/cfg"
	"github.com/zew/https-server/pkg/gziphandler"
	"github.com/zew/https-server/pkg/gzipper"
)

// serviceWorkerPreCache contains settings
// to include directory files into service worker install pre-cache
type serviceWorkerPreCache struct {
	cache             bool     // part of service worker install pre-cache
	includeExtensions []string // only include files with these extensions, i.e. ".webp"
}

// dirT contains settings for a directory of static files
type dirT struct {
	isSingleFile bool // mostly directories, a few special files like favicon.ico, robots.txt

	src      string // server side directory name; app path
	srcTpl   string // server side template file
	urlPath  string // client side URI
	MimeType string

	includeExtensions []string // only include files with these extensions, i.e. ".webp"

	preGZIP  bool // file is gzipped during preprocess
	liveGZIP bool // file is gzipped at each request time
	swpc     serviceWorkerPreCache

	HeadTemplate string // template for HTML head section, i.e. "<script src=..." or "<link href=..."

	HTTPCache int // HTTP response header caching time
}

type dirsT []dirT

func (dirs dirsT) Register(mux *http.ServeMux) {

	for _, dir := range dirs {

		if dir.isSingleFile {
			mux.HandleFunc(dir.urlPath, serveStatic)
		} else {
			if cfg.Get().StaticFilesGZIP && dir.preGZIP {
				mux.HandleFunc(dir.urlPath, serveStatic)
			} else if cfg.Get().StaticFilesGZIP && dir.liveGZIP {
				mux.Handle(dir.urlPath, gziphandler.GzipHandler(http.HandlerFunc(serveStatic)))
			} else {
				mux.HandleFunc(dir.urlPath, serveStatic)
				// mux.Handle(dir.urlPath, http.HandlerFunc(serveStatic))
			}
		}

	}

}

func Register(mux *http.ServeMux) {
	dirs.Register(mux)
}

var dirs = dirsT{

	// special files - single files
	{
		isSingleFile: true,
		urlPath:      "/service-worker.js",
		src:          "./app-bucket/js-service-worker/service-worker.js",
		srcTpl:       "./app-bucket/tpl/service-worker.tpl.js",
		MimeType:     "application/javascript",

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: false,
		},
		HTTPCache: 60 * 60 * 120,
	},
	{
		isSingleFile: true,
		// always requested from root
		urlPath:  "/favicon.ico",
		src:      "./app-bucket/img/favicon.ico",
		MimeType: "image/x-icon",

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		HTTPCache: 60 * 60 * 120,
	},
	{
		isSingleFile: true,
		urlPath:      "/robots.txt",
		src:          "./app-bucket/txt/robots.txt",
		MimeType:     "text/plain",
		// no caching since file is not requested by web browsers
	},

	//
	// directories
	{
		src:      "./app-bucket/js/",
		urlPath:  "/js/",
		MimeType: "application/javascript",

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		HeadTemplate: `	<script src="/js/%s/%v" nonce="%s" ></script>`,
		HTTPCache: 60 * 60 * 120,
	},
	{
		src:      "./app-bucket/css/",
		urlPath:  "/css/",
		MimeType: "text/css",

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		HeadTemplate: `	<link href="/css/%s/%s" nonce="%s" rel="stylesheet" type="text/css" media="screen" />`,
		HTTPCache: 60 * 60 * 120,
	},
	{
		src:      "./app-bucket/json/",
		urlPath:  "/json/",
		MimeType: "application/json",

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		// HeadTemplate: ``,
		HTTPCache: 60 * 60 * 120,
	},
	{
		src:               "./app-bucket/img/",
		urlPath:           "/img/",
		MimeType:          "image/webp",
		includeExtensions: []string{".webp", ".ico"},

		preGZIP: false, // webp files are already compressed
		swpc: serviceWorkerPreCache{
			cache:             true,
			includeExtensions: []string{".webp"},
		},
		HTTPCache: 60 * 60 * 120,
	},
}

func init() {
	// cfg.RunHandleFunc(prepareStatic, "/prepare-static")
}

// filesOfDir returns all files and directories inside dirSrc
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

/*
	a clean way to create a JavaScript file
	from a template file
	with JavaScript fields  {{ .MyJavaScriptSnippet  }}

	we dont want "<" being escaped to &lt; in the template

	and we dont want "'" in .MyJavaScriptSnippet being escaped to &23232;

*/
func javascriptTemplate(srcPth string, w io.Writer) {

	_ = texttpl.HTMLEscapeString("force-import")
	_ = htmltpl.HTMLEscapeString("force-import")

	btsSW, err := os.ReadFile(srcPth)
	if err != nil {
		fmt.Fprintf(w, "could not read service worker template %v\n", err)
		return
	}
	strSW := string(btsSW)
	strSWAsJS := htmltpl.JS(strSW)
	_ = strSWAsJS

	tpl2 := htmltpl.New("xx")
	_ = tpl2
	tpl2.Parse("{{js .JS}}")

}

func serviceWorkerCacheList(w http.ResponseWriter) []string {

	var staticFilesSeed = []string{
		"/index.html",
		"/offline.html",
		// ... dynamic css, js, webp, json

		// example for external url res
		// "https://fonts.google.com/icon?family=Material+Icons",
	}

	staticFiles := make([]string, len(staticFilesSeed), 16)
	copy(staticFiles, staticFilesSeed)

	for _, dir := range dirs {

		if dir.isSingleFile {
			if !dir.swpc.cache {
				continue
			}

			// 	these files are requested without
			//  at version prefix in the path
			//  thus the caching also does not need the cache
			staticFiles = append(staticFiles, dir.urlPath)
			continue
		}

		files, err := filesOfDir(dir.src)
		if err != nil {
			fmt.Fprint(w, err)
			return staticFiles
		}

		for _, file := range files {

			if file.IsDir() {
				continue
			}

			ext := path.Ext(file.Name())
			if ext != ".webp" &&
				ext != ".css" &&
				ext != ".js" &&
				ext != ".json" {
				continue
			}

			// fmt.Fprintf(w, "\t %v\n", file.Name())
			pth := path.Join(dir.src, cfg.Get().TS, file.Name())
			pth = strings.Replace(pth, "app-bucket/", "./", 1)
			staticFiles = append(staticFiles, pth)
		}
	}

	return staticFiles
}

// prepareServiceWorker compiles a list of files and a version
func prepareServiceWorker(w http.ResponseWriter, req *http.Request) {

	srcPth := "./app-bucket/tpl/service-worker.tpl.js"
	t, err := htmltpl.ParseFiles(srcPth)
	if err != nil {
		fmt.Fprintf(w, "could not parse service worker template %v\n", err)
		return
	}

	bts := &bytes.Buffer{}

	data := struct {
		Version     string
		ListOfFiles htmltpl.HTML // neither .JS nor .JSStr do work
	}{
		Version:     cfg.Get().TS,
		ListOfFiles: htmltpl.HTML("'" + strings.Join(serviceWorkerCacheList(w), "',\n  '") + "'"),
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
func PrepareStatic(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "prepare static files\n")

	prepareServiceWorker(w, req)

	//
	//
	dirs := []string{
		"./app-bucket/js/",
		"./app-bucket/js-service-worker/",
		"./app-bucket/css/",
		"./app-bucket/json/",
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
				// media="screen" to prevent lighthouse from flagging resource as blocking
				fmt.Fprintf(sb, `	<link href="/css/%s/%s" nonce="%s" rel="stylesheet" type="text/css" media="screen" />`, cfg.Get().TS, file.Name(), cfg.Get().TS)
				fmt.Fprint(sb, "\n")
			}

			fnSrc := path.Join(dirSrc, file.Name())
			fnDst := path.Join(dirDst, file.Name())

			// gzip JavaScript and CSS, but not images
			if dirIdx < 4 {
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
			if dirIdx == 4 {
				ext := path.Ext(file.Name())
				// favicon is read from here
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

// addVersion purpose
//      if images or json files are requested without any version
// 		then serve current version
func addVersion(r *http.Request) *http.Request {

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 3 { // behold empty token from leading "/"
		return r
	}
	// log.Printf("parts are %+v", parts)
	needle := fmt.Sprintf("/%v/", parts[1])
	replacement := fmt.Sprintf("/%v/%v/", parts[1], cfg.Get().TS)
	r.URL.Path = strings.Replace(r.URL.Path, needle, replacement, 1)
	// log.Printf("new URL  %v", r.URL.Path)
	return r
}

func serveStatic(w http.ResponseWriter, r *http.Request) {

	switch {

	//
	// special files
	case strings.HasPrefix(r.URL.Path, "/robots.txt"):
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	case strings.HasPrefix(r.URL.Path, "/favicon.ico"):
		// browsers request this from the root
		w.Header().Set("Content-Type", "image/x-icon")
		r.URL.Path = fmt.Sprintf("/img/%v/favicon.ico", cfg.Get().TS)
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", 60*60*120))

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
		r = addVersion(r)
	case strings.HasPrefix(r.URL.Path, "/json/"):
		w.Header().Set("Content-Type", "application/json")
		r = addVersion(r)
	default:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "unsupported static directory")
		return
	}

	pth := "./app-bucket" + r.URL.Path

	if cfg.Get().StaticFilesGZIP {
		// not for images, not for json
		if strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/js-service-worker/") ||
			strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/json/") {
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
