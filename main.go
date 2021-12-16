package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/zew/https-server/cfg"
	"github.com/zew/https-server/gziphandler"
	"golang.org/x/crypto/acme/autocert"
)

func stripPortFromHost(s string) string {
	if strings.Contains(s, ":") {
		return strings.Split(s, ":")[0]
	}
	return s
}

func redirectHTTP(w http.ResponseWriter, r *http.Request) {

	if cfg.Get().AutoRedirectHTTP {

		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Use HTTPS", http.StatusBadRequest)
			return
		}
		target := "https://" + stripPortFromHost(r.Host) + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusFound)

	} else {

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "You asked for HTTP - please switch to HTTPS.\n")

	}

}

func renderScaffold(w http.ResponseWriter, title, desc string, cnt *strings.Builder) {

	w.Header().Set("Content-Type", "text/html")
	/*
		https://csp.withgoogle.com/docs/index.html
		https://csp.withgoogle.com/docs/strict-csp.html#example

		default-src     -  https: http:
		script-src      - 'nonce-{random}'  'unsafe-inline'   - for old browsers
		style-src-elem  - 'self' 'nonce-{random}'             - strangely, self is required
		strict-dynamic  -  unused; 'script-xyz.js'can load additional scripts via script elements
		object-src      -  sources for <object>, <embed>, <applet>
		base-uri        - 'self' 'none' - for relative URLs
		report-uri      -  https://your-report-collector.example.com/

	*/
	csp := ""
	csp += fmt.Sprintf("default-src     https:; ")
	csp += fmt.Sprintf("base-uri       'none'; ")
	csp += fmt.Sprintf("object-src     'none'; ")
	csp += fmt.Sprintf("script-src     'nonce-%d' 'unsafe-inline'; ", cfg.Get().TS)
	csp += fmt.Sprintf("style-src-elem 'nonce-%d' 'self'; ", cfg.Get().TS)
	csp += fmt.Sprintf("worker-src      https://*/service-worker.js; ")
	w.Header().Set("Content-Security-Policy", csp)

	fmt.Fprintf(w, cfg.Get().HTML5, title, desc, cfg.Get().JS, cfg.Get().CSS, cnt.String())
}

func home(w http.ResponseWriter, r *http.Request) {

	cnt := &strings.Builder{}
	fmt.Fprintf(cnt, "<p>Hello, TLS user </p>\n")
	fmt.Fprintf(cnt, "<p>Your config: </p>\n")
	fmt.Fprintf(cnt, "<pre id='tls-config'>%+v  </pre>\n", r.TLS)

	renderScaffold(w, "Home page", "golang web server prototype", cnt)

}

func plain(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "This is an example server.\n")
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/hello", plain)

	mux.Handle("/js/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.Handle("/css/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.Handle("/img/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.Handle("/json/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))

	// special static files - must be in root dir
	mux.HandleFunc("/robots.txt", staticResources)
	mux.HandleFunc("/service-worker.js", staticResources)

	var err error

	switch cfg.Get().ModeHTTPS {

	case "https-localhost-cert":
		// localhost development; based on [Filipo Valsordas tool](https://github.com/FiloSottile/mkcert)
		err = http.ListenAndServeTLS(
			":443", "./certs/server.pem", "./certs/server.key", mux)

	case "letsenrypt-simple":
		lstnr := autocert.NewListener(cfg.Get().Domains...)
		err = http.Serve(lstnr, mux)

	case "letsenrypt-extended":

		mgr := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Get().Domains...),
			Cache:      autocert.DirCache(filepath.Join(cfg.Get().AppDir, "autocert")),
		}

		tlsCfg := &tls.Config{
			// ServerName:     "xxxxxx",  // should equal hostname
			GetCertificate: mgr.GetCertificate,
			MinVersion:     tls.VersionTLS12,
			// NextProtos: []string{
			// 	"h2", "http/1.1", // enable HTTP/2
			// 	acme.ALPNProto, // enable tls-alpn ACME challenges
			// },
		}

		server := &http.Server{
			Addr:      ":https", // same as :443
			TLSConfig: tlsCfg,
			Handler:   mux,
		}

		if false {
			// this would disable http/2
			server.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}

		// http server
		go func() {
			if cfg.Get().AllowHTTP {
				log.Fatal(http.ListenAndServe(":http", mux))
			} else {
				log.Printf("serve HTTP, which will redirect automatically to HTTPS - %v", cfg.Get().Dms)
				// h := mgr.HTTPHandler(nil) // argument nil would silently redirect
				h := mgr.HTTPHandler(http.HandlerFunc(redirectHTTP))
				log.Fatal(http.ListenAndServe(":http", h))
			}
		}()

		log.Printf("serve HTTPS for %v", cfg.Get().Dms)
		err = server.ListenAndServeTLS("", "")

	}

	if err != nil {
		log.Fatal("https server failed: ", err)
	}
}
