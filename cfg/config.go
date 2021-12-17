package cfg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type cfg struct {
	ModeHTTPS string

	Domains []string
	Dms     string // string representation of Domains

	AllowHTTP        bool // or forward to HTTPS, only implemented for ModeHTTPS==letsenrypt-extended
	AutoRedirectHTTP bool // when HTTP is not allowed: Redirect automatically redirect to HTTPS via web browser

	HTML5 string
	JS    string
	CSS   string

	AppDir string
	TS     int64

	// serve precompressed .gzip files
	//   default is dynamic compression
	PrecompressedGZIP bool

	// default app title, meta tag description
}

var defaultCfg = &cfg{
	ModeHTTPS: "https-localhost-cert",
	// ModeHTTPS: "letsenrypt-simple",
	// ModeHTTPS: "letsenrypt-extended",

	Domains: []string{"fmt.zew.de", "fmt.zew.de"},

	//
	AllowHTTP:        true,
	AutoRedirectHTTP: true,

	PrecompressedGZIP: true,
}

func init() {

	log.SetFlags(log.Lshortfile | log.Llongfile)

	bts, _ := json.MarshalIndent(defaultCfg, "", "\t")
	ioutil.WriteFile("tmp-example-config.json", bts, 0777)

	pth := "./config.json"
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
		pth := "./scaffold.html"
		bts, err := ioutil.ReadFile(pth)
		if err != nil {
			log.Printf("error opening %v, %v", pth, err)
			return
		}
		defaultCfg.HTML5 = string(bts)
	}
}

func Get() *cfg {
	return defaultCfg
}
