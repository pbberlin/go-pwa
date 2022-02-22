package static

import "net/http"

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
	"/only-template-preparation-1": {
		isSingleFile: true,
		srcTpl:       "./app-bucket/tpl/db.tpl.js",
		tplExecutor:  execDB,

		src: "./app-bucket/js/",
		fn:  "db.js",
	},
	"/only-template-preparation-2": {
		isSingleFile: true,
		srcTpl:       "./app-bucket/tpl/manifest.tpl.json",
		tplExecutor:  execManifest,

		src: "./app-bucket/json/",
		fn:  "/manifest.json",
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
