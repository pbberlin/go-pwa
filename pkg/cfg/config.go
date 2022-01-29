// package cfg load app scope JSON settings
// its init() func needs to be executed first
package cfg

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

// Headless executes a HTTP handle func
// and writes it output into the log
func Headless(fnc func(w http.ResponseWriter, r *http.Request), pth string) {
	log.SetFlags(log.Lshortfile | log.Llongfile)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", pth, nil)
	fnc(w, r)

	res := w.Result()
	defer res.Body.Close()
	bts, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response of %v: %v", pth, err)
	}
	log.Print(string(bts))
}

func init() {
	// RunHandleFunc(Load, "/config/load")
}

//
// cfg contain app scope data
type cfg struct {
	Title       string `json:"title,omitempty"`       // default app title
	Description string `json:"description,omitempty"` // default meta tag description

	ModeHTTPS string `json:"mode_https,omitempty"`

	Domains []string `json:"domains,omitempty"`

	AllowHTTP        bool `json:"allow_http,omitempty"`         // or forward to HTTPS, only implemented for ModeHTTPS==letsenrypt-extended
	AutoRedirectHTTP bool `json:"auto_redirect_http,omitempty"` // when HTTP is not allowed: Redirect automatically redirect to HTTPS via web browser

	// serve precompressed .gzip files
	//   default is dynamic compression via gziphandler
	PrecompressGZIP bool `json:"precompress_gzip,omitempty"`

	//
	// dynamically set on app init
	TplMain *template.Template `json:"-"` // loaded from scaffold.tpl.html

	CSP string `json:"-"` // content security policy
	JS  string `json:"-"` // created from ./app-bucket/js
	CSS string `json:"-"` // created from ./app-bucket/css

	Dms string `json:"-"` // string representation of domains

	AppDir string `json:"-"` // server side app dir; do we still need this - or is ./... always sufficient?
	TS     string `json:"-"` // prefix 'vs-'  and then timestamp of app start

	PrefURI string `json:"pref_uri,omitempty"`
}

var defaultCfg = &cfg{
	Title:       "your app title",
	Description: "default meta tag description",

	ModeHTTPS: "https-localhost-cert",
	// ModeHTTPS: "letsenrypt-simple",
	// ModeHTTPS: "letsenrypt-extended",

	Domains: []string{"fmt.zew.de", "fmt.zew.de"},

	//
	AllowHTTP:        true,
	AutoRedirectHTTP: true,

	PrecompressGZIP: true,
}

func (cfg *cfg) cSP() string {

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
	csp += fmt.Sprintf("script-src     'nonce-%s' 'unsafe-inline'; ", cfg.TS)
	csp += fmt.Sprintf("style-src-elem 'nonce-%s' 'self'; ", cfg.TS)
	csp += fmt.Sprintf("worker-src      https://*/service-worker.js; ")

	return csp
}

func Load(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "config load start\n")

	// write example with defaults
	{
		pth := "./app-bucket/server-config/tmp-example-config.json"
		bts, err := json.MarshalIndent(defaultCfg, "", "\t")
		if err != nil {
			fmt.Fprintf(w, "error marshalling defaultCfg %v \n", err)
		}

		err = os.WriteFile(pth, bts, 0777)
		if err != nil {
			fmt.Fprintf(w, "error saving %v, %v \n", pth, err)
		}
	}

	//
	tmpCfg := cfg{}
	pth := "./app-bucket/server-config/config.json"
	bts, err := os.ReadFile(pth)
	if err != nil {
		fmt.Fprintf(w, "error opening %v, %v \n", pth, err)
		panic(fmt.Sprintf("need %v", pth))
	} else {
		err := json.Unmarshal(bts, &tmpCfg)
		if err != nil {
			fmt.Fprintf(w, "error unmarshalling %v, %v \n", pth, err)
			panic(fmt.Sprintf("need valid %v", pth))
		} else {
			fmt.Fprintf(w, "config loaded from %v \n", pth)
		}
	}

	//
	//
	// other dynamic stuff
	tmpCfg.Dms = strings.Join(tmpCfg.Domains, ", ")

	tmpCfg.AppDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("error os.Getwd() %v \n", err)
	}

	// time stamp
	// keeping only the last 6 digits
	now := time.Now().UTC().Unix()
	fst4 := float64(now) / float64(1000*1000) // first four digits
	fst4 = math.Floor(fst4) * float64(1000*1000)
	now -= int64(fst4)

	tmpCfg.TS = fmt.Sprintf("vs-%d", now)

	tmpCfg.CSP = tmpCfg.cSP()

	//
	{
		pth := "./app-bucket/tpl/scaffold.tpl.html"
		var err error
		tmpCfg.TplMain, err = template.ParseFiles(pth)
		if err != nil {
			log.Printf("error parsing  %v, %v", pth, err)
			return
		}
	}

	defaultCfg = &tmpCfg // replace pointer at once - should be threadsafe

	fmt.Fprint(w, "config load stop\n\n")

}

func Get() *cfg {
	return defaultCfg
}

func Set() *cfg {
	return defaultCfg
}
