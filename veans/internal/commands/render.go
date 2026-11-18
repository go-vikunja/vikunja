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

package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
)

// fieldSet is the parsed --fields value; nil prints whole tasks.
type fieldSet []string

var taskFields = func() []string {
	var names []string
	typ := reflect.TypeFor[client.Task]()
	for i := range typ.NumField() {
		name, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ",")
		if name != "" && name != "-" {
			names = append(names, name)
		}
	}
	return names
}()

func addFieldsFlag(cmd *cobra.Command, raw *string) {
	cmd.Flags().StringVar(raw, "fields", "",
		"comma-separated task fields to print, e.g. identifier,url (default: all)")
}

// parseFields runs before any request so a typo can't follow a mutation.
func parseFields(raw string) (fieldSet, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var fs fieldSet
	for f := range strings.SplitSeq(raw, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if !slices.Contains(taskFields, f) {
			return nil, output.New(output.CodeValidation,
				"unknown field %q in --fields (known: %s)", f, strings.Join(taskFields, ", "))
		}
		fs = append(fs, f)
	}
	return fs, nil
}

// prepareTask fills URLs and drops nested related-task descriptions, which
// bloat the output and are one `veans show` away.
func (r *runtime) prepareTask(t *client.Task) {
	t.URL = r.cfg.TaskURL(t.ID)
	for _, related := range t.RelatedTasks {
		for _, rt := range related {
			if rt == nil {
				continue
			}
			rt.URL = r.cfg.TaskURL(rt.ID)
			rt.Description = ""
		}
	}
}

func (r *runtime) printTask(w io.Writer, fs fieldSet, t *client.Task) error {
	r.prepareTask(t)
	v, err := fs.project(t)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(v)
}

func (r *runtime) printTasks(w io.Writer, fs fieldSet, tasks []*client.Task) error {
	out := make([]any, 0, len(tasks))
	for _, t := range tasks {
		r.prepareTask(t)
		v, err := fs.project(t)
		if err != nil {
			return err
		}
		out = append(out, v)
	}
	return json.NewEncoder(w).Encode(out)
}

// project keeps the requested keys in the requested order; a field the
// task omits (omitempty) prints as null so callers see a stable shape.
func (fs fieldSet) project(t *client.Task) (any, error) {
	if fs == nil {
		return t, nil
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range fs {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(f)) // json tag names, plain ASCII
		buf.WriteByte(':')
		if v, ok := all[f]; ok {
			buf.Write(v)
		} else {
			buf.WriteString("null")
		}
	}
	buf.WriteByte('}')
	return json.RawMessage(buf.Bytes()), nil
}
