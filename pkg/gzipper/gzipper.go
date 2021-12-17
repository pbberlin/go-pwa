// package gzipper compresses a file to file.gzip
package gzipper

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type gzFileWriter struct {
	f   *os.File // underlying file
	gzW *gzip.Writer
	buf *bufio.Writer // on top of gzW
}

func New(fn string) (gzFileWriter, error) {
	if !strings.HasSuffix(fn, ".gzip") {
		fn += ".gzip"
	}
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return gzFileWriter{}, fmt.Errorf("error creating gzip base %v, %v", fn, err)
	}
	gzW := gzip.NewWriter(f)
	buf := bufio.NewWriter(gzW)
	return gzFileWriter{f, gzW, buf}, nil
}

func (gz gzFileWriter) Write(bts []byte) (int, error) {
	return (gz.buf).Write(bts)
}

func (gz gzFileWriter) WriteString(s string) (int, error) {
	return (gz.buf).WriteString(s)
}

func (gz gzFileWriter) WriteFile(srcPth string) {

	src, err := os.Open(srcPth)
	if err != nil {
		log.Printf("error opening gzip source file %v, %v", srcPth, err)
		return
	}
	defer src.Close()

	n, err := io.Copy(gz.buf, src)
	if err != nil {
		log.Printf("error writing file into gzip file: %v, %v", srcPth, err)
		return
	}

	log.Printf("%8v bytes written to %v from %v", n, gz.f.Name(), srcPth)

}

func (gz gzFileWriter) Close() {
	gz.buf.Flush()
	gz.gzW.Close() // gzip first
	gz.f.Close()   // file last
}
