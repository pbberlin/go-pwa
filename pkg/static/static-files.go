/* package static - preparing and serving static files.
It assumes a directory ./app-bucket containing directories by mime-types.
/css, /js, /img...
The package also supports typical special files: robots.txt, service-worker.js, favicon.ico
being served under special URIs.
The package takes care of
  * template execution
  * mime typing
  * HTTP caching
  * service worker pre-caching
  * consistent versioning
  * gzipping
  * registering routes with a http.ServeMux
  * handler funcs for HTTP request serving

Template execution allows custom funcs for arbitrary dynamic preparations.

A few are provided, to prepare Google PWA config files
dynamically from whatever is in the directories under ./app-bucket.

We dynamically generate
* a list of files for the service worker pre-cache
* a list of icons for manifest.json
* a version constant for indexed DB schema version in db.js

All file preparation logic is put together in the HTTP handle func PrepareStatic(...).
Thus you whenever you changed any static file contents,
call PrepareStatic(), and you get a *consistent* new version of all static files,
and you force your HTTP client (aka browser) to load

Todo:
* Make the config loadable via JSON
* Javascript templating is done in a highly inappropriate way; cannot get idiomatic way to work
* Markdown with some pre-processing is missing

*/
package static

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/zew/https-server/pkg/cfg"
	"github.com/zew/https-server/pkg/gziphandler"
	"github.com/zew/https-server/pkg/gzipper"
)

func init() {
	// cfg.RunHandleFunc(prepareStatic, "/prepare-static")
}

// serviceWorkerPreCache contains settings
// to include directory files into service worker install pre-cache
type serviceWorkerPreCache struct {
	cache             bool     // part of service worker install pre-cache
	includeExtensions []string // only include files with these extensions, i.e. ".webp"
}

// dirT contains settings for a directory of static files
type dirT struct {
	isSingleFile bool // mostly directories, a few special files like favicon.ico, robots.txt

	srcTpl      string // template file to generate src in pre-pre run
	tplExecutor func(dirsT, dirT, http.ResponseWriter)

	src string // server side directory name; app path
	fn  string // file name - for single files

	urlPath   string // client side URI
	MimeType  string
	HTTPCache int // HTTP response header caching time

	includeExtensions []string // only include files with these extensions into directory processing, i.e. ".webp"

	preGZIP  bool // file is gzipped during preprocess
	liveGZIP bool // file is gzipped at each request time
	swpc     serviceWorkerPreCache

	HeadTemplate string // template for HTML head section, i.e. "<script src=..." or "<link href=..."

}

type dirsT map[string]dirT

// filesOfDir is a helper returning all files and directories inside dirSrc
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

func zipOrCopy(w http.ResponseWriter, dir dirT, fn string) {

	if dir.isSingleFile {
		fn = dir.fn
	}

	dirDst := path.Join(dir.src, fmt.Sprint(cfg.Get().TS))
	err := os.MkdirAll(dirDst, 0755)
	if err != nil {
		fmt.Fprintf(w, "could not create dir %v, %v\n", dir.src, err)
		return
	}

	pthSrc := path.Join(dir.src, fn)
	pthDst := path.Join(dirDst, fn)

	// gzip JavaScript and CSS, but not images
	if dir.preGZIP {
		//
		// closure because of defer
		zipIt := func() {
			gzw, err := gzipper.New(pthDst)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
			defer gzw.Close()
			// gzw.WriteString("gzipper.go created this file.\n")
			err = gzw.WriteFile(pthSrc)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		}
		zipIt()
	} else {
		// dont gzip
		bts, err := os.ReadFile(pthSrc)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		os.WriteFile(pthDst, bts, 0644)
	}

	fmt.Fprintf(w, "\t %v\n", fn)

}

// PrepareStatic creates directories with versioned files
// PrepareStatic zips JS and CSS files
// PrepareStatic compiles "<script src..." and "<link href="/" blocks
func (dirs dirsT) PrepareStatic(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "prepare static files\n")

	for _, dir := range dirs {

		if dir.isSingleFile {

			// if dir.srcTpl != "" {
			// 	dir.execServiceWorker(w)
			if dir.tplExecutor != nil {
				dir.tplExecutor(dirs, dir, w)
			}
			zipOrCopy(w, dir, "")

			continue // separate loop
		}

		fmt.Fprintf(w, "start src dir %v\n", dir.src)

		sb := &strings.Builder{}

		files, err := filesOfDir(dir.src)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		for _, file := range files {

			if strings.HasSuffix(file.Name(), ".gzip") {
				continue
			}

			if len(dir.includeExtensions) > 0 {
				for _, ext := range dir.includeExtensions {
					if strings.HasSuffix(file.Name(), ext) {
						continue
					}
				}
			}

			if file.IsDir() {
				continue
			}

			if dir.HeadTemplate != "" {
				// fmt.Fprintf(sb, `	<script src="/js/%s/%v" nonce="%s" ></script>`, cfg.Get().TS, file.Name(), cfg.Get().TS)
				fmt.Fprintf(sb, dir.HeadTemplate, cfg.Get().TS, file.Name(), cfg.Get().TS)
				fmt.Fprint(sb, "\n")
			}

			zipOrCopy(w, dir, file.Name())

		}

		if strings.HasPrefix(dir.urlPath, "/js/") {
			cfg.Set().JS = sb.String()
		}
		if strings.HasPrefix(dir.urlPath, "/css/") {
			cfg.Set().CSS = sb.String()
		}

		fmt.Fprintf(w, "stop src dir  %v\n\n", dir.src)

	}

}

// addVersion is a helper to serveStatic; its purpose:
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

func (dirs dirsT) serveStatic(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/") // avoid empty first token with split below
	prefs := strings.Split(path, "/")
	pref := "/" + prefs[0]
	// log.Printf("static server path %v - prefix %v", path, pref)

	dir, ok := dirs[pref]
	if ok {
		if dir.isSingleFile {
			// log.Printf("static server path %v - prefix %v", path, pref)
		}
		// r = addVersion(r)
		if dir.MimeType != "" {
			w.Header().Set("Content-Type", dir.MimeType)
		}
		if dir.HTTPCache > 0 {
			w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", dir.HTTPCache))
		}

	} else {
		log.Printf("no static server info for %v", pref)
		fmt.Fprintf(w, "no static server info for %v", pref)
		return
	}

	if dir.isSingleFile {
		// r.URL.Path = fmt.Sprintf("/txt/%v/robots.txt", cfg.Get().TS)
		// r.URL.Path = fmt.Sprintf("/img/%v/favicon.ico", cfg.Get().TS)
		// r.URL.Path = fmt.Sprintf("/js-service-worker/%v/service-worker.js", cfg.Get().TS)

		r.URL.Path = fmt.Sprintf("%v%v/%v", dir.src, cfg.Get().TS, dir.fn)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "./app-bucket")
	}

	pth := "./app-bucket" + r.URL.Path

	if cfg.Get().StaticFilesGZIP && dir.preGZIP {
		// everything but images
		// is supposed to be pre-compressed

		// if
		//  HasPrefix(r.URL.Path, "/js/") ||
		// 	HasPrefix(r.URL.Path, "/css/") || ...
		pth += ".gzip"
		w.Header().Set("Content-Encoding", "gzip")
	}

	file, err := os.Open(pth)
	if err != nil {
		log.Printf("error: %v", err) // path is in error msg
		referrer := r.Header.Get("Referer")
		if referrer != "" {
			log.Printf("\tfrom referrer %+v", referrer)
		}
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

// Register URL handlers
func (dirs dirsT) Register(mux *http.ServeMux) {

	for _, dir := range dirs {
		if dir.isSingleFile {
			mux.HandleFunc(dir.urlPath, dirs.serveStatic)
		} else {
			if cfg.Get().StaticFilesGZIP && dir.preGZIP {
				mux.HandleFunc(dir.urlPath, dirs.serveStatic)
			} else if cfg.Get().StaticFilesGZIP && dir.liveGZIP {
				mux.Handle(dir.urlPath, gziphandler.GzipHandler(http.HandlerFunc(dirs.serveStatic)))
			} else {
				mux.HandleFunc(dir.urlPath, dirs.serveStatic)
				// mux.Handle(dir.urlPath, http.HandlerFunc(dirs.serveStatic))
			}
		}
	}

}

//
//
// default static config

var staticDirsDefault = dirsT{

	// special files - single files
	"/service-worker.js": {
		isSingleFile: true,
		srcTpl:       "./app-bucket/tpl/service-worker.tpl.js",
		tplExecutor:  execServiceWorker,

		src: "./app-bucket/js-service-worker/",
		fn:  "service-worker.js",

		urlPath:   "/service-worker.js",
		MimeType:  "application/javascript",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: false,
		},
	},
	"/manifest.json": {
		isSingleFile: true,
		srcTpl:       "./app-bucket/tpl/manifest.tpl.json",
		tplExecutor:  execManifest,

		src: "./app-bucket/json/",
		fn:  "/manifest.json",

		urlPath:   "/manifest.json",
		MimeType:  "application/json",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
	},
	"/favicon.ico": {
		isSingleFile: true,
		// always requested from root
		src: "./app-bucket/img/",
		fn:  "favicon.ico",

		urlPath:   "/favicon.ico",
		MimeType:  "image/x-icon",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
	},
	"/robots.txt": {
		isSingleFile: true,
		src:          "./app-bucket/txt/",
		fn:           "robots.txt",

		urlPath:  "/robots.txt",
		MimeType: "text/plain",
		// no caching since file is not requested by web browsers
	},

	//
	// directories
	"/js": {
		src: "./app-bucket/js/",

		urlPath:   "/js/",
		MimeType:  "application/javascript",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		HeadTemplate: `	<script src="/js/%s/%v" nonce="%s" ></script>`,
	},
	"/css": {
		src: "./app-bucket/css/",

		urlPath:   "/css/",
		MimeType:  "text/css",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		HeadTemplate: `	<link href="/css/%s/%s" nonce="%s" rel="stylesheet" type="text/css" media="screen" />`,
	},
	"/json": {
		src: "./app-bucket/json/",

		urlPath:   "/json/",
		MimeType:  "application/json",
		HTTPCache: 60 * 60 * 120,

		preGZIP: true,
		swpc: serviceWorkerPreCache{
			cache: true,
		},
		// HeadTemplate: ``,
	},
	"/img": {
		src: "./app-bucket/img/",

		urlPath:           "/img/",
		MimeType:          "image/webp",
		HTTPCache:         60 * 60 * 120,
		includeExtensions: []string{".webp", ".ico"},

		preGZIP: false, // webp files are already compressed
		swpc: serviceWorkerPreCache{
			cache:             true,
			includeExtensions: []string{".webp"},
		},
	},
}

func Register(mux *http.ServeMux) {
	staticDirsDefault.Register(mux)
}

func PrepareStatic(w http.ResponseWriter, req *http.Request) {
	staticDirsDefault.PrepareStatic(w, req)
}
