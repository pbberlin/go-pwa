package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/zew/https-server/pkg/cfg"
	"github.com/zew/https-server/pkg/gziphandler"
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

func main() {

	var err error

	cfg.Headless(cfg.Load, "/config/load")
	cfg.Headless(prepareStatic, "/prepare-static")

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/index.html", home)

	// mux.HandleFunc("/prepare-static", prepareStatic)
	// mux.HandleFunc("/config/load", cfg.Load)
	mux.HandleFunc("/config/load", func(w http.ResponseWriter, r *http.Request) {
		cfg.Load(w, r)
		prepareStatic(w, r)
	})
	mux.HandleFunc("/hello", plain)

	if cfg.Get().PrecompressGZIP {
		mux.HandleFunc("/js/", staticResources)
		mux.HandleFunc("/css/", staticResources)
	} else {
		mux.Handle("/js/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
		mux.Handle("/css/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	}

	mux.Handle("/json/", http.HandlerFunc(staticResources))
	mux.Handle("/img/", http.HandlerFunc(staticResources))

	// special static files - must be in root dir
	mux.HandleFunc("/robots.txt", staticResources)
	mux.HandleFunc("/service-worker.js", staticResources)

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
