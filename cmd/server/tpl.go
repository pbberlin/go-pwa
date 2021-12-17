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

func (sc *Scaffold) render(w http.ResponseWriter, cnt *strings.Builder) {

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
