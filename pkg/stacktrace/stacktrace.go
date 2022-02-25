package stacktrace

import (
	"log"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"unicode/utf8"
)

var (
	appDir          string // ~work dir
	homeDirPackages string
	goRoot          string
	goModCache      string
)

func init() {
	var err error
	appDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("could not determine working dir: %v", err)
	}
	appDir = strings.ReplaceAll(appDir, "\\", "/") // stack trace is always forward slash - even on windows
	log.Printf("workdir is %v", appDir)

	homeDir, err := os.UserHomeDir() // for instance C:/Users/pbu/
	if err != nil {
		log.Fatalf("could not determine user home dir: %v", err)
	}
	homeDir = strings.ReplaceAll(homeDir, "\\", "/") // stack trace is always forward slash - even on windows
	log.Printf("homeDir is %v", homeDir)

	homeDirPackages = path.Join(homeDir, "go", "pkg", "src")
	log.Printf("homeDirPackages is %v", homeDirPackages)

	goRoot = runtime.GOROOT()                      // i.e. C:\Program Files\Go
	goRoot = strings.ReplaceAll(goRoot, "\\", "/") // stack trace is always forward slash - even on windows
	goRoot = path.Join(goRoot, "src")
	log.Printf("goRoot is %v", goRoot)

	goModCache = os.Getenv("GOMODCACHE")                   // i.e. C:\Users\pbu\go\pkg\mod
	goModCache = strings.ReplaceAll(goModCache, "\\", "/") // stack trace is always forward slash - even on windows
	if goModCache == "" {
		goModCache = path.Join(homeDir, "go", "pkg")
	}
	log.Printf("goModCache is %v", goModCache)

}

func removeLastRune(str string) string {
	for len(str) > 0 {
		_, size := utf8.DecodeLastRuneInString(str)
		return str[:len(str)-size]
	}
	return str
}

func cleanse(s string) string {

	if strings.HasPrefix(s, "  ") || strings.HasPrefix(s, "\t") {
		// code location line => remove ugly strange line ending chars
		pos := strings.Index(s, "+0x")
		if pos > -1 {
			s = s[:pos]
		}
		// s = removeLastRune(s) + "--"
	} else {
		// func name line => throw away ugly pointers of arguments
		pos := strings.Index(s, "(")
		if pos > -1 {
			s = s[:pos] + "(..."
		}

	}
	s = strings.ReplaceAll(s, homeDirPackages, " HDP: ")
	s = strings.ReplaceAll(s, appDir, "   APP: ")
	s = strings.ReplaceAll(s, goRoot, "   GOR: ")
	s = strings.ReplaceAll(s, goModCache, "   GMC: ")

	return s
}

func Log() {

	st := string(debug.Stack())
	sts := strings.Split(st, "\n")

	// cut off first lines for this func itself
	sts = sts[7:]
	if sts[len(sts)-1] == "" {
		sts = sts[:len(sts)-1]
	}

	// chop off leading panic entry
	if strings.HasPrefix(sts[0], "panic(") {
		sts = sts[2:]
	}

	// chop off boring first entry
	ln1 := len(sts)
	if strings.HasPrefix(sts[ln1-2], "created by net/http.(") {
		sts = sts[:len(sts)-2]
	}

	// chop off entire head of http server stuff; contains no info
	for rowIdx, row := range sts {
		if strings.HasPrefix(row, "net/http.(") {
			sts = sts[:rowIdx]
			break
		}
	}

	for i := 0; i < len(sts); i++ {
		// log.Printf("%2v: %v", i+1, cleanse(sts[i]))
		log.Printf("%v", cleanse(sts[i]))
	}

}
