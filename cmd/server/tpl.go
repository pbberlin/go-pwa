package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/zew/https-server/pkg/cfg"
)

type Scaffold struct {
	Title, Desc string

	JSTags  template.HTML
	CSSTags template.HTML
	Content template.HTML
}

// SP provides the static path for a uri;
// prepending an app prefix and a version
func (sc *Scaffold) SP(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 3 { // behold empty token from leading "/"
		return uri
	}
	needle := fmt.Sprintf("/%v/", parts[1])
	replacement := fmt.Sprintf("/%v/%v/", parts[1], cfg.Get().TS)
	uri = strings.Replace(uri, needle, replacement, 1)
	return uri
}

func (sc *Scaffold) render(w http.ResponseWriter, cnt *strings.Builder) {

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Security-Policy", cfg.Get().CSP)

	if sc.Title == "" {
		sc.Title = cfg.Get().Title
	}
	if sc.Desc == "" {
		sc.Desc = cfg.Get().Description
	}

	sc.JSTags = template.HTML(cfg.Get().JS)
	sc.CSSTags = template.HTML(cfg.Get().CSS)

	sc.Content = template.HTML(cnt.String())

	// standard output to print merged data
	err := cfg.Get().TplMain.Execute(w, sc)
	if err != nil {
		fmt.Fprintf(w, "error executing tplMain: %v", err)
		return
	}

	// fmt.Fprintf(w, cfg.Get().HTML5, sc.Title, sc.Desc, cfg.Get().JS, cfg.Get().CSS, cnt.String())
}
