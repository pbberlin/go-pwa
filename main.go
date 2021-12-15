package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

const rewriteHTTP = true
const forwardHTTP = false

func redirectHTTP(w http.ResponseWriter, r *http.Request) {

	if rewriteHTTP {

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

// description is an requirement by Google's lighthouse HTML testing suite
var html5 = ``

func init() {
	pth := "./scaffold.html"
	bts, err := ioutil.ReadFile(pth)
	if err != nil {
		log.Printf("error opening %v, %v", pth, err)
		return
	}
	html5 = string(bts)
}

// close by using </body></html>

var js = `
	<script src="/js/menu-and-form-control-keys.js?v=1637681051" ></script>
	<script src="/js/validation.js?v=1637681051"                 ></script>
`
var css = `
	<link href="/css/styles.css?v=1637681051"              rel="stylesheet" type="text/css" />
	<link href="/css/progress-bar-2.css?v=1637681051"      rel="stylesheet" type="text/css" />
	<link href="/css/styles-mobile.css?v=1637681051"       rel="stylesheet" type="text/css" />
	<link href="/css/styles-quest.css?v=1637681051"        rel="stylesheet" type="text/css" />
	<link href="/css/styles-quest-no-site-specified-a.css" rel="stylesheet" type="text/css" />
`

func home(w http.ResponseWriter, r *http.Request) {

	// cnt := strings.Builder{}
	cnt := &bytes.Buffer{}
	fmt.Fprintf(cnt, "<p>Hello, TLS user </p>\n")
	fmt.Fprintf(cnt, "<p>Your config: </p>\n")
	fmt.Fprintf(cnt, "<pre  style='white-space: pre-wrap;'>%+v  </pre>\n", r.TLS)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, html5, "Hello, TLS user", js, css, cnt.String())

}

func plainHello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "This is an example server.\n")
}

func staticResources(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/js/") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasPrefix(r.URL.Path, "/css/") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasPrefix(r.URL.Path, "/robots.txt") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User-agent: Googlebot Disallow: /example-subfolder/")
		return
	} else {
		return
	}
	pth := "./static" + r.URL.Path

	// bts, _ := ioutil.ReadFile(pth)
	// fmt.Fprint(w, bts)

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

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/hello", plainHello)

	mux.Handle("/js/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.Handle("/css/", gziphandler.GzipHandler(http.HandlerFunc(staticResources)))
	mux.HandleFunc("/robots.txt", staticResources)

	var err error

	switch "letsenrypt-extended" {

	case "classic":
		err = http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)

	case "letsenrypt-simple":
		lstnr := autocert.NewListener("fmt.zew.de")
		err = http.Serve(lstnr, mux)

	case "letsenrypt-extended":
		appDir, _ := os.Getwd()
		domains := []string{"fmt.zew.de", "fmt.zew.de"}
		// domains := []string{"fmt.zew.de", "dmd.zew.de"}
		domainss := strings.Join(domains, ", ")

		mgr := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
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

			if forwardHTTP {
				log.Printf("serve HTTP, which will redirect automatically to HTTPS - %v", domainss)
				// h := mgr.HTTPHandler(nil) // argument nil would silently redirect
				h := mgr.HTTPHandler(http.HandlerFunc(redirectHTTP))
				log.Fatal(http.ListenAndServe(":http", h))
			} else {
				// allow http
				log.Fatal(http.ListenAndServe(":http", mux))
			}
		}()

		log.Printf("serve HTTPS for %v", domainss)
		err = server.ListenAndServeTLS("", "")

	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
