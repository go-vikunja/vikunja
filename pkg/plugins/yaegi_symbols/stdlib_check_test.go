package yaegi_symbols

import (
	"testing"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func TestYaegiSmoke(t *testing.T) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	_, err := i.Eval(`import "fmt"`)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	v, err := i.Eval(`fmt.Sprintf("hello %s", "yaegi")`)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if v.String() != "hello yaegi" {
		t.Fatalf("unexpected result: %s", v.String())
	}
}
