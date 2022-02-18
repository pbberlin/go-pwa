package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/zew/https-server/pkg/cfg"
	"github.com/zew/https-server/pkg/db"
	"github.com/zew/https-server/pkg/static"
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
	cfg.Headless(static.PrepareStatic, "/prepare-static")

	db.Initialize()

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/index.html", home)
	mux.HandleFunc("/offline.html", offline)

	// mux.HandleFunc("/prepare-static", prepareStatic)
	// mux.HandleFunc("/config/load", cfg.Load)
	mux.HandleFunc("/config/load", func(w http.ResponseWriter, r *http.Request) {
		cfg.Load(w, r)
		static.PrepareStatic(w, r)
	})
	mux.HandleFunc("/hello", plain)
	mux.HandleFunc("/save-json", saveJson)

	static.Register(mux)

	switch cfg.Get().ModeHTTPS {

	case "https-localhost-cert":
		// localhost development using https://github.com/FiloSottile/mkcert
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
