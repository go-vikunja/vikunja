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

package vikunjafile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
)

// encoding/json rejects deeper documents anyway; the cap keeps the recursion off the stack limit.
const maxDataDepth = 10000

// Room for uncounted per-item data such as a position and a bucket in each of ~25 views plus reactions;
// CheckItemLimit is the exact cap.
const decodeBytesPerItem = 8 << 10

// Charged for every element or key whose decode target can't be resolved; exceeds the largest export struct.
const unknownElementBytes = 1 << 10

var unmarshalerType = reflect.TypeFor[json.Unmarshaler]()

// Streams data against the decode target's types so the heap json.Unmarshal would allocate stays bounded.
func checkDataItemCount(data []byte) error {
	b := &decodeBudget{
		dec:       json.NewDecoder(bytes.NewReader(data)),
		remaining: migration.MaxItems() * decodeBytesPerItem,
	}
	err := b.value(reflect.TypeFor[[]*models.ProjectWithTasksAndBuckets](), 0)
	var limitErr *migration.ErrImportItemLimitExceeded
	if err == nil || errors.As(err, &limitErr) {
		return err
	}
	return &migration.ErrInvalidImportFile{Err: fmt.Errorf("could not read data: %w", err)}
}

type decodeBudget struct {
	dec       *json.Decoder
	remaining int64
}

func (b *decodeBudget) charge(size uintptr) error {
	b.remaining -= int64(size) // #nosec G115 - type sizes are tiny
	if b.remaining < 0 {
		return &migration.ErrImportItemLimitExceeded{MaxItems: migration.MaxItems()}
	}
	return nil
}

// t is what json.Unmarshal decodes the next value into, or nil when that isn't known exactly.
func (b *decodeBudget) value(t reflect.Type, depth int) error {
	tok, err := b.dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		return nil
	}
	for t != nil && t.Kind() == reflect.Pointer {
		if err := b.charge(t.Elem().Size()); err != nil {
			return err
		}
		t = t.Elem()
	}
	if t != nil && (t.Kind() == reflect.Interface || reflect.PointerTo(t).Implements(unmarshalerType)) {
		t = nil
	}
	delim, isContainer := tok.(json.Delim)
	if !isContainer {
		return nil
	}
	if depth >= maxDataDepth {
		return errors.New("data nested too deeply")
	}

	var elem reflect.Type
	var elemCost uintptr = unknownElementBytes
	var fields map[string]reflect.Type
	switch {
	case t == nil:
	case delim == '[' && t.Kind() == reflect.Slice:
		// Doubled for append's spare capacity.
		elem, elemCost = t.Elem(), 2*t.Elem().Size()
	case delim == '{' && t.Kind() == reflect.Map:
		elem, elemCost = t.Elem(), 3*(t.Key().Size()+t.Elem().Size())
	case delim == '{' && t.Kind() == reflect.Struct:
		fields = jsonFields(t)
	}

	for b.dec.More() {
		next, cost := elem, elemCost
		if delim == '{' {
			key, err := b.dec.Token()
			if err != nil {
				return err
			}
			if fields != nil {
				// A miss may still case-fold onto a field in encoding/json, so it's charged as unknown.
				var ok bool
				if next, ok = fields[key.(string)]; ok {
					cost = 0
				}
			}
		}
		if err := b.charge(cost); err != nil {
			return err
		}
		if err := b.value(next, depth+1); err != nil {
			return err
		}
	}
	_, err = b.dec.Token()
	return err
}

var jsonFieldsCache sync.Map // reflect.Type -> map[string]reflect.Type

// Exact json names of t's fields as encoding/json resolves them; ambiguous names are left out.
func jsonFields(t reflect.Type) map[string]reflect.Type {
	if cached, ok := jsonFieldsCache.Load(t); ok {
		return cached.(map[string]reflect.Type)
	}
	fields := map[string]reflect.Type{}
	depths := map[string]int{}
	var collect func(t reflect.Type, depth int)
	collect = func(t reflect.Type, depth int) {
		for i := range t.NumField() {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			// Embedded pointers are left out: json would allocate them, so their fields fall back to unknown.
			if f.Anonymous && name == "" && f.Type.Kind() == reflect.Struct {
				collect(f.Type, depth+1)
				continue
			}
			if !f.IsExported() {
				continue
			}
			if name == "" {
				name = f.Name
			}
			switch d, seen := depths[name]; {
			case !seen || depth < d:
				fields[name], depths[name] = f.Type, depth
			case depth == d:
				delete(fields, name)
			}
		}
	}
	collect(t, 0)
	jsonFieldsCache.Store(t, fields)
	return fields
}
