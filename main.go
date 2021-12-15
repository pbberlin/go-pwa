package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zew/https-server/cfg"
	"github.com/zew/https-server/gziphandler"
	"golang.org/x/crypto/acme"
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

func home(w http.ResponseWriter, r *http.Request) {

	cnt := &strings.Builder{}
	fmt.Fprintf(cnt, "<p>Hello, TLS user </p>\n")
	fmt.Fprintf(cnt, "<p>Your config: </p>\n")
	fmt.Fprintf(cnt, "<pre id='tls-config'>%+v  </pre>\n", r.TLS)

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
	csp := fmt.Sprintf("default-src https:; script-src 'nonce-%d' 'unsafe-inline'; style-src-elem 'self' 'nonce-%d'; object-src 'none'; base-uri 'none';", cfg.Get().TS, cfg.Get().TS)
	w.Header().Set("Content-Security-Policy", csp)
	fmt.Fprintf(w, cfg.Get().HTML5, "Hello, TLS user", js, css, cnt.String())

}

func plainHello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "This is an example server.\n")
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/hello", plainHello)

	mux.Handle("/js/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.Handle("/css/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.HandleFunc("/robots.txt", staticResources)

	var err error

	switch cfg.Get().ModeHTTPS {

	case "https-localhost-cert":
		err = http.ListenAndServeTLS(
			":443", "./certs/server.pem", "./certs/server.key", mux)

	case "letsenrypt-simple":
		lstnr := autocert.NewListener(cfg.Get().Domains...)
		err = http.Serve(lstnr, mux)

	case "letsenrypt-extended":
		appDir, _ := os.Getwd()

		mgr := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Get().Domains...),
			Cache:      autocert.DirCache(filepath.Join(appDir, "autocert")),
		}

		tlsCfg := &tls.Config{
			GetCertificate: mgr.GetCertificate,
			NextProtos: []string{
				"h2", "http/1.1", // enable HTTP/2
				acme.ALPNProto, // enable tls-alpn ACME challenges
			},

			// only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519, // Go 1.8 only
			},
			MinVersion: tls.VersionTLS12,

			/*
				// For TLS 1.2 - Default: client ciphersuite preferences,
				// switchting to server ciphersuite prefs of Ciphersuites
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
			*/
		}

		server := &http.Server{
			Addr:      ":https",
			TLSConfig: tlsCfg,
			Handler:   mux,
		}

		if false {
			// this would disable http/2
			server.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}

		// http fallback
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
		log.Fatal("ListenAndServe: ", err)
	}
}
