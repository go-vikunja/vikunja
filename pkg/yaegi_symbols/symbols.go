package yaegi_symbols

import (
	"reflect"

	"code.vikunja.io/api/pkg/log"
)

// Symbols contains all Vikunja and third-party symbols for use with yaegi's Use() method.
var Symbols = map[string]map[string]reflect.Value{}

// yaegi extract treats Fatal* in any package named "log" as restricted and
// references these local wrappers instead of the package functions, so they
// must exist for the generated vikunja_log.go to compile.
func logFatal(v ...interface{})            { log.Fatal(v...) }
func logFatalf(f string, v ...interface{}) { log.Fatalf(f, v...) }
