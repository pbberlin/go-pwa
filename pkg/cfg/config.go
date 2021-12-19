// package cfg load app scope JSON settings
// its init() func needs to be executed first
package cfg

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Llongfile)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/config/load", nil)
	Load(w, r)

	res := w.Result()
	defer res.Body.Close()
	bts, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response of /config/load %v", err)
	}
	log.Print(string(bts))
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

	JS  string `json:"-"` // created from ./static/js
	CSS string `json:"-"` // created from ./static/css

	Dms string `json:"-"` // string representation of domains

	AppDir string `json:"-"` // do we still need this - or is ./... always sufficient?
	TS     int64  `json:"-"` // timestamp of app start
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

func Load(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "loading config\n")

	{
		pth := "./static/json/tmp-example-config.json"
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
	pth := "./static/json/config.json"
	bts, err := os.ReadFile(pth)
	if err != nil {
		fmt.Fprintf(w, "error opening %v, %v \n", pth, err)
		panic(fmt.Sprintf("need %v", pth))
	} else {
		err := json.Unmarshal(bts, defaultCfg)
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
	defaultCfg.Dms = strings.Join(defaultCfg.Domains, ", ")

	defaultCfg.AppDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("error os.Getwd() %v \n", err)
	}

	defaultCfg.TS = time.Now().UTC().Unix()

	//
	{
		pth := "./static/tpl/scaffold.tpl.html"
		var err error
		defaultCfg.TplMain, err = template.ParseFiles(pth)
		if err != nil {
			log.Printf("error parsing  %v, %v", pth, err)
			return
		}
	}

}

func Get() *cfg {
	return defaultCfg
}

func Set() *cfg {
	return defaultCfg
}
