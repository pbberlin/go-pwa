// package gzipper compresses a file to file.gzip
package gzipper

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

type gzFileWriter struct {
	f   *os.File // underlying file
	gzW *gzip.Writer
	buf *bufio.Writer // on top of gzW
}

func New(fnDst string) (gzFileWriter, error) {
	if !strings.HasSuffix(fnDst, ".gzip") {
		fnDst += ".gzip"
	}
	f, err := os.OpenFile(fnDst, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return gzFileWriter{}, fmt.Errorf("error creating gzip destination %v, %v", fnDst, err)
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

func (gz gzFileWriter) WriteFile(srcPath string) error {

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening gzip src file %v, %w", srcPath, err)
	}
	defer src.Close()

	n, err := io.Copy(gz.buf, src)
	_ = n
	if err != nil {
		return fmt.Errorf("error writing filecontents of %v into %v, %w", srcPath, gz.f.Name(), err)
	}

	// log.Printf("%8v bytes written to %v from %v", n, gz.f.Name(), srcPth)
	return nil

}

func (gz gzFileWriter) Close() {
	gz.buf.Flush()
	gz.gzW.Close() // gzip first
	gz.f.Close()   // file last
}
