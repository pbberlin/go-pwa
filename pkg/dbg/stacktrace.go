// package dbg contains a de-cluttered debug.Stack();
// and object dump to JSON string
package dbg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"unicode/utf8"
)

// for removing standard paths such as /c/program files/go...
// from the trace
var (
	appDir          string // ~work dir
	homeDirPackages string // /home/username/go/pkg/src
	goRoot          string // /c/program files/go
	goModCache      string // similar to homeDirPackages
)

var DEBUG = false

func packageLogger(fmt string, s ...string) {
	if DEBUG {
		log.Printf(fmt, s)
	}
}

func init() {

	var err error
	appDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("could not determine working dir: %v", err)
	}
	appDir = strings.ReplaceAll(appDir, "\\", "/") // stack trace is always forward slash - even on windows
	packageLogger("workdir is %v", appDir)

	homeDir, err := os.UserHomeDir() // for instance C:/Users/pbu/
	if err != nil {
		log.Fatalf("could not determine user home dir: %v", err)
	}
	homeDir = strings.ReplaceAll(homeDir, "\\", "/") // stack trace is always forward slash - even on windows
	packageLogger("homeDir is %v", homeDir)

	homeDirPackages = path.Join(homeDir, "go", "pkg", "src")
	packageLogger("homeDirPackages is %v", homeDirPackages)

	goRoot = runtime.GOROOT()                      // i.e. C:\Program Files\Go
	goRoot = strings.ReplaceAll(goRoot, "\\", "/") // stack trace is always forward slash - even on windows
	goRoot = path.Join(goRoot, "src")
	packageLogger("goRoot is %v", goRoot)

	goModCache = os.Getenv("GOMODCACHE")                   // i.e. C:\Users\pbu\go\pkg\mod
	goModCache = strings.ReplaceAll(goModCache, "\\", "/") // stack trace is always forward slash - even on windows
	if goModCache == "" {
		goModCache = path.Join(homeDir, "go", "pkg", "src")
	}
	packageLogger("goModCache is %v", goModCache)

}

// unused
func removeLastRune(str string) string {
	for len(str) > 0 {
		_, size := utf8.DecodeLastRuneInString(str)
		return str[:len(str)-size]
	}
	return str
}

// cleanse removes clutter such as long, repetitive paths and mere pointer addresses
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

// prepare chops off boilerplate lines from the end
// and from the start of the stack
func prepare() []string {

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

	return sts
}

// StackTracePre returns stacktrace for HTML with an optional CSS format string
func StackTracePre(styles ...string) string {

	style := ""
	if len(styles) > 0 {
		style = styles[0]
	}

	sts := prepare()
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "<pre style='%v'>\n", style)
	for i := 0; i < len(sts); i++ {
		fmt.Fprintf(sb, "%v\n", cleanse(sts[i]))
	}
	fmt.Fprint(sb, "</pre>\n")
	return sb.String()

}

// StackTrace prints stacktrace to the standard logger
func StackTrace() {

	sts := prepare()
	for i := 0; i < len(sts); i++ {
		// log.Printf("%2v: %v", i+1, cleanse(sts[i]))
		log.Printf("%v", cleanse(sts[i]))
	}

}

//
//
// ---------------------
//
//

// Dump2String like Dump, but JSON string is returned
func Dump2String(v interface{}) string {

	sb := &strings.Builder{}

	if false {
		var strValue reflect.Value = reflect.ValueOf(&v) // take pointer
		indirectStr := reflect.Indirect(strValue)        // pointer back to value
		typeAsString := reflect.MakeMap(indirectStr.Type())
		fmt.Fprintf(sb, "\ttype:  %v\n", typeAsString)
	}
	fmt.Fprintf(sb, "\n\ttype  %T:\n", v)

	firstColLeftMostPrefix := " "
	bts, err := json.MarshalIndent(v, firstColLeftMostPrefix, "\t")
	if err != nil {
		s := fmt.Sprintf("error indent: %v\n", err)
		return s
	}

	bts = bytes.Replace(bts, []byte(`\u003c`), []byte("<"), -1)
	bts = bytes.Replace(bts, []byte(`\u003e`), []byte(">"), -1)
	bts = bytes.Replace(bts, []byte(`\n`), []byte("\n"), -1)

	fmt.Fprint(sb, string(bts))

	return sb.String()
}

// Dump converts to JSON string and prints to standard logger;
// sort of Dump2Log
func Dump(v interface{}) {
	log.Print(Dump2String(v))
}

// Dump2Pre like Dump, , but JSON string is returned nested in <pre>...</pre>
func Dump2Pre(v interface{}, styles ...string) string {

	style := ""
	if len(styles) > 0 {
		style = styles[0]
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "<pre style='%v'>\n", style)
	fmt.Fprint(sb, Dump2String(v))
	fmt.Fprint(sb, "</pre>\n")
	return sb.String()

}
