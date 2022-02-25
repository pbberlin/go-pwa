package stacktrace

import (
	"log"
	"runtime/debug"
	"strings"
	"unicode/utf8"
)

func RemoveLastChar(str string) string {
	for len(str) > 0 {
		_, size := utf8.DecodeLastRuneInString(str)
		return str[:len(str)-size]
	}
	return str
}

func Log() {

	cleanse := func(s string) string {

		if !strings.HasPrefix(s, " ") {
			pos := strings.Index(s, "(")
			if pos > -1 {
				s = s[:pos] + "(..."
			}
		} else {
			pos := strings.Index(s, "+0x")
			if pos > -1 {
				log.Printf("found +0x in %v", s)
				s = s[:pos]
			} else {
				s = RemoveLastChar(s)
			}
		}

		s = strings.ReplaceAll(s, "C:/Users/pbu/Documents/zew_work/git/go/", "   ")
		s = strings.ReplaceAll(s, "C:/Program Files/Go/", "   ")

		return s
	}

	st := string(debug.Stack())
	sts := strings.Split(st, "\n")

	sts = sts[7:] // cut off first lines for this func itself
	if sts[len(sts)-1] == "" {
		sts = sts[:len(sts)-1]
	}
	if strings.HasPrefix(sts[0], "panic(") {
		sts = sts[2:]
	}

	maxRows := 10

	ln := len(sts)
	if ln > maxRows {
		ln = maxRows
	}

	for i := 0; i < ln; i++ {
		log.Printf("%2v: %v", i, cleanse(sts[i]))
	}
	log.Print("   ...")
	for i := len(sts) - ln; i < len(sts); i++ {
		log.Printf("%2v: %v", i, cleanse(sts[i]))
	}

}
