// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestAsError_Nil(t *testing.T) {
	if got := AsError(nil); got != nil {
		t.Fatalf("AsError(nil) = %#v, want nil", got)
	}
}

func TestAsError_PreservesKnownCode(t *testing.T) {
	orig := New(CodeValidation, "bad input")
	got := AsError(orig)
	if got != orig {
		t.Fatalf("AsError returned a different pointer; want the original *Error to be preserved via errors.As")
	}
	if got.Code != CodeValidation {
		t.Fatalf("Code = %q, want %q", got.Code, CodeValidation)
	}
	if got.Message != "bad input" {
		t.Fatalf("Message = %q, want %q", got.Message, "bad input")
	}
}

func TestAsError_UnwrapsThroughFmtErrorf(t *testing.T) {
	inner := New(CodeNotFound, "missing")
	wrapped := fmt.Errorf("context: %w", inner)
	got := AsError(wrapped)
	if got != inner {
		t.Fatalf("AsError did not return the inner *Error through fmt.Errorf wrapping")
	}
	if got.Code != CodeNotFound {
		t.Fatalf("Code = %q, want %q", got.Code, CodeNotFound)
	}
}

func TestAsError_PlainErrorBecomesUnknown(t *testing.T) {
	plain := errors.New("kaboom")
	got := AsError(plain)
	if got == nil {
		t.Fatal("AsError(plain) = nil, want a CodeUnknown wrapper")
	}
	if got.Code != CodeUnknown {
		t.Fatalf("Code = %q, want %q", got.Code, CodeUnknown)
	}
	if got.Message != "kaboom" {
		t.Fatalf("Message = %q, want %q", got.Message, "kaboom")
	}
	if !errors.Is(got, plain) {
		t.Fatal("CodeUnknown wrapper does not preserve the original cause via Unwrap")
	}
}

func TestEmitError_EnvelopeShape(t *testing.T) {
	var buf bytes.Buffer
	EmitError(&Error{Code: CodeValidation, Message: "x"}, &buf)

	out := buf.Bytes()
	// json.Encoder.Encode appends a trailing newline; assert and tolerate it.
	if len(out) == 0 || out[len(out)-1] != '\n' {
		t.Fatalf("expected trailing newline from json.Encoder.Encode, got %q", string(out))
	}

	// Decode into a generic map first to assert the exact key set.
	var asAny map[string]any
	if err := json.Unmarshal(out, &asAny); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, string(out))
	}
	if len(asAny) != 2 {
		t.Fatalf("envelope has %d keys, want exactly 2 (code, error); got %v", len(asAny), asAny)
	}
	if _, ok := asAny["code"]; !ok {
		t.Fatalf("envelope missing %q key; got %v", "code", asAny)
	}
	if _, ok := asAny["error"]; !ok {
		t.Fatalf("envelope missing %q key; got %v", "error", asAny)
	}

	// And confirm both fields decode as strings with the expected values.
	var asStrings map[string]string
	if err := json.Unmarshal(out, &asStrings); err != nil {
		t.Fatalf("envelope fields are not all strings: %v (%q)", err, string(out))
	}
	if asStrings["code"] != string(CodeValidation) {
		t.Fatalf("code = %q, want %q", asStrings["code"], string(CodeValidation))
	}
	if asStrings["error"] != "x" {
		t.Fatalf("error = %q, want %q", asStrings["error"], "x")
	}
}

// NOTE: EmitError's fallback path (the fmt.Fprintf to os.Stderr when
// json.Encoder.Encode fails) is intentionally not unit-tested. The encoded
// value is an *Error with two string fields and no custom MarshalJSON, so
// json.Marshal cannot fail on it from outside the package — there is no
// stdlib-reachable input that trips the encoder. The fallback's contract
// ("preserve the {code,error} envelope shape even on encode failure") is
// covered by inspection of the source: the format string emits the same two
// keys with CodeUnknown and a descriptive message.

func TestWrap_PreservesCauseForErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel cause")
	wrapped := Wrap(CodeConflict, sentinel, "while doing thing %d", 42)

	if !errors.Is(wrapped, sentinel) {
		t.Fatal("errors.Is(Wrap(...), sentinel) = false, want true; Wrap must preserve the cause through Unwrap")
	}

	// errors.As against the sentinel's concrete type should also walk the
	// chain; use a custom type to make this meaningful.
	custom := &causeType{msg: "custom"}
	wrapped2 := Wrap(CodeConflict, custom, "wrap")
	var target *causeType
	if !errors.As(wrapped2, &target) {
		t.Fatal("errors.As did not find the wrapped cause through the *Error chain")
	}
	if target != custom {
		t.Fatalf("errors.As returned a different pointer than the original cause")
	}

	// And the wrapped *Error itself still carries the supplied code and
	// formatted message.
	if wrapped.Code != CodeConflict {
		t.Fatalf("Code = %q, want %q", wrapped.Code, CodeConflict)
	}
	if wrapped.Message != "while doing thing 42" {
		t.Fatalf("Message = %q, want %q", wrapped.Message, "while doing thing 42")
	}
}

type causeType struct{ msg string }

func (c *causeType) Error() string { return c.msg }
