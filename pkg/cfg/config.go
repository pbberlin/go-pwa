package cfg

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

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

	AppDir string `json:"-"`
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

func init() {

	log.SetFlags(log.Lshortfile | log.Llongfile)

	bts, _ := json.MarshalIndent(defaultCfg, "", "\t")
	ioutil.WriteFile("./static/json/tmp-example-config.json", bts, 0777)

	//
	pth := "./static/json/config.json"
	bts, err := ioutil.ReadFile(pth)
	if err != nil {
		log.Printf("error opening %v, %v", pth, err)
	} else {
		err := json.Unmarshal(bts, defaultCfg)
		if err != nil {
			log.Printf("error unmarshalling %v, %v", pth, err)
		} else {
			log.Printf("config loaded from %v", pth)
		}
	}

	defaultCfg.Dms = strings.Join(defaultCfg.Domains, ", ")

	defaultCfg.AppDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("error os.Getwd() %v", err)
	}

	defaultCfg.TS = time.Now().UTC().Unix()

	//
	{
		pth := "./static/tpl/scaffold.tpl.html"
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
