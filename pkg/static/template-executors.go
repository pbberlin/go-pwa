package static

import (
	"bytes"
	"fmt"
	"html/template"
	htmltpl "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	texttpl "text/template"

	"github.com/zew/https-server/pkg/cfg"
)

// listForPreCaching compiles the list of files with a version
// to be pre-cached by service worker
func (dirs dirsT) listForPreCaching(w http.ResponseWriter) []string {

	var seed = []string{
		"/index.html",
		"/offline.html",
		// ... dynamic css, js, webp, json

		// example for external url res
		// "https://fonts.google.com/icon?family=Material+Icons",
	}

	res := make([]string, len(seed), 16)
	copy(res, seed)

	for _, dir := range dirs {

		if dir.isSingleFile {
			if !dir.swpc.cache {
				continue
			}
			// 	these files are requested without
			//  version prefix in the path,
			//  thus we dont need the version prefix here
			res = append(res, dir.urlPath)
			continue
		}

		files, err := filesOfDir(dir.src)
		if err != nil {
			fmt.Fprint(w, err)
			return res
		}

		for _, file := range files {

			if file.IsDir() {
				continue
			}

			if len(dir.swpc.includeExtensions) > 0 {
				for _, ext := range dir.swpc.includeExtensions {
					if strings.HasSuffix(file.Name(), ext) {
						continue
					}
				}
			}

			// fmt.Fprintf(w, "\t %v\n", file.Name())
			pth := path.Join(dir.src, cfg.Get().TS, file.Name())
			pth = strings.Replace(pth, "app-bucket/", "./", 1)
			res = append(res, pth)
		}
	}

	return res
}

/*
	This is unused and leads nowhere.

	There is no idiomatic way to fill a template of Javascript
	with snippets of Javascript :-(

	A clean way woudl be to create a template file
	from a JavaScript file fill its
	JavaScript fields  {{ .MyJavaScriptSnippet  }}

	We would want the template be of type html.JS or html.JSStr

	If we use HTML types, we get escapings

		"<" being escaped to &lt; in the template

		"'" in .MyJavaScriptSnippet being escaped to &23232;


*/
func javascriptTemplate(srcPth string, w io.Writer) {

	_ = htmltpl.HTMLEscapeString("force-import")
	_ = texttpl.HTMLEscapeString("force-import") // avoiding escaping hassle by using text templates => fail

	btsSW, err := os.ReadFile(srcPth)
	if err != nil {
		fmt.Fprintf(w, "could not read service worker template %v\n", err)
		return
	}
	strSW := string(btsSW)         // file content bytes to string
	strSWAsJS := htmltpl.JS(strSW) // string into type Javascript code
	_ = strSWAsJS

	//

	tpl := htmltpl.New("JS body")
	tpl.Parse("let x=1; {{js .JS}}") // we want to put in strSWAsJS, but the argument must be string

}

// execServiceWorker executes the service worker template
// and writes the result into the file system;
// see javascriptTemplate for a discussion of alternatives
func execServiceWorker(dirs dirsT, dir dirT, w http.ResponseWriter) {

	t, err := htmltpl.ParseFiles(dir.srcTpl)
	if err != nil {
		fmt.Fprintf(w, "could not parse template %v: %v\n", dir.srcTpl, err)
		return
	}

	data := struct {
		Version     string
		ListOfFiles htmltpl.HTML // neither .JS nor .JSStr do work
	}{
		Version:     cfg.Get().TS,
		ListOfFiles: htmltpl.HTML("'" + strings.Join(dirs.listForPreCaching(w), "',\n  '") + "'"),
	}

	bts := &bytes.Buffer{}
	err = t.Execute(bts, data)
	if err != nil {
		fmt.Fprintf(w, "could not execute template %v: %v\n", dir.srcTpl, err)
		return
	}

	dstPth := path.Join(dir.src, dir.fn)
	os.WriteFile(dstPth, bts.Bytes(), 0644)
}

func manifestIconList(w http.ResponseWriter) string {

	/* template string for each file
		{
			"src": "/img/icon-072.webp",
			"sizes": "72x72",
			"type": "image/png",
			"purpose": "maskable"
	  	},
	*/
	ts := `  {
    "src": "/img/{{.FN}}",
    "sizes": "{{.Size}}",
    "type": "image/png",
    "purpose": "{{.MaskableOrAny}}"
  }`

	t := template.New("list-entry")
	t, err := t.Parse(ts)
	if err != nil {
		log.Fatalf("error parsing icon template: %v", err)
	}

	// globbing for .../img/icon-072.webp, ... /img/icon-128.webp ...
	icons, err := filepath.Glob("./app-bucket/img/icon-???.webp")
	if err != nil {
		log.Fatalf("error reading icons: %v", err)
	}

	sb := &strings.Builder{}

	for idx, icon := range icons {
		icon = filepath.Base(icon)
		log.Printf("\ticon %v", icon)

		fnp := icon                                       // file name part containing the icon size, i.e. 072
		fnp = strings.TrimSuffix(fnp, filepath.Ext(icon)) // remove trailing .webp
		fnp = strings.TrimPrefix(fnp, "icon-")            // leading  icon-
		if strings.HasPrefix(fnp, "0") {
			fnp = strings.TrimPrefix(fnp, "0")
		}
		// size := "072x072"
		size := fmt.Sprintf("%vx%v", fnp, fnp)

		maskableOrAny := "maskable"
		if len(fnp) > 2 && fnp > "128" {
			maskableOrAny = "any"
		}

		data := struct {
			FN            string
			Size          string
			MaskableOrAny string
		}{
			FN:            icon,
			Size:          size,
			MaskableOrAny: maskableOrAny,
		}
		t.Execute(sb, data)
		if idx < len(icons)-1 {
			fmt.Fprint(sb, ",\n") // trailing comma - except for last iteration
		}
	}

	return sb.String()
}

func execManifest(dirs dirsT, dir dirT, w http.ResponseWriter) {

	t, err := htmltpl.ParseFiles(dir.srcTpl)
	if err != nil {
		fmt.Fprintf(w, "could not parse template %v: %v\n", dir.srcTpl, err)
		return
	}

	data := struct {
		Title       string
		TitleShort  string
		Description string
		IconList    htmltpl.HTML
	}{
		Title:       cfg.Get().Title,
		TitleShort:  cfg.Get().TitleShort,
		Description: cfg.Get().Description,
		IconList:    htmltpl.HTML(manifestIconList(w)),
	}

	bts := &bytes.Buffer{}
	err = t.Execute(bts, data)
	if err != nil {
		fmt.Fprintf(w, "could not execute template %v: %v\n", dir.srcTpl, err)
		return
	}

	dstPth := path.Join(dir.src, dir.fn)
	os.WriteFile(dstPth, bts.Bytes(), 0644)
}
