package static

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pbberlin/go-pwa/pkg/cfg"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Llongfile)
}

func TestMainGetH(t *testing.T) {
	test(t, "GET")
}

func TestMainPostH(t *testing.T) {
	// test(t, "POST")
}

func test(t *testing.T, method string) {

	//
	// app init stuff
	_ = os.Chdir("../..")
	wd, _ := os.Getwd()
	t.Logf("working dir is %v", wd)
	cfg.Headless(cfg.Load, "/config/load")         // create version
	cfg.Headless(PrepareStatic, "/prepare-static") // create versioned files

	//
	//
	pth := fmt.Sprint("/favicon.ico")
	postBody := &bytes.Buffer{}

	req, err := http.NewRequest(method, pth, postBody) // <-- encoded payload
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Accept", "image/x-icon")

	// http.Client

	w := httptest.NewRecorder() //  satisfying http.ResponseWriter for recording
	handler := http.HandlerFunc(staticDirsDefault.serveStatic)

	handler.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("%v: status code: got %v, want %v", pth, status, http.StatusOK)
	} else {
		t.Logf("%v: Response code OK", pth)
	}

	/* 	rsp := w.Result()
	   	got, err := io.ReadAll(rsp.Body)
	   	if err != nil {
	   		t.Errorf("ReadAll failed: %v", err)
	   	}
	*/
	gotZipped := w.Body.Bytes()

	rdr, err := gzip.NewReader(bytes.NewBuffer(gotZipped))
	if err != nil {
		t.Logf("could not read the response as gzip: %v", err)
	}
	defer rdr.Close()

	got, err := io.ReadAll(rdr)
	if err != nil {
		t.Errorf("ReadAll failed: %v", err)
	}

	// compare the response body
	pth2 := "./app-bucket/img/favicon.ico"
	t.Logf("Reading from %v", pth2)
	wnt, err := os.ReadFile(pth2)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(got, wnt) {
		t.Errorf("%v: handler returned wrong img -  got %v Bytes -- wnt %v Bytes", pth, len(got), len(wnt))
		for key, hdr := range w.HeaderMap {
			t.Logf("    %-20v - %+v", key, hdr)
		}
		// ioutil.WriteFile(fmt.Sprintf("tmp-test_%v_want.html", method), []byte(expected), 0777)
		// ioutil.WriteFile(fmt.Sprintf("tmp-test_%v_got.html", method), []byte(body), 0777)
	}
}
