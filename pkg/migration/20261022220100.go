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

package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/log"

	"github.com/ganigeorgiev/fexpr"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// savedFilter20261022220100 reads saved_filters.filters as raw JSON text, so the
// migration controls the (un)marshalling and keeps keys it does not know.
type savedFilter20261022220100 struct {
	ID      int64  `xorm:"'id'"`
	Filters string `xorm:"'filters'"`
}

func (savedFilter20261022220100) TableName() string {
	return "saved_filters"
}

// projectView20261022220100 reads the two JSON filter columns of project_views as
// raw text. NULL reads as "".
type projectView20261022220100 struct {
	ID                  int64  `xorm:"'id'"`
	Filter              string `xorm:"'filter'"`
	BucketConfiguration string `xorm:"'bucket_configuration'"`
}

func (projectView20261022220100) TableName() string {
	return "project_views"
}

const (
	repeatAfter20261022220100 = "repeat_after"
	repeatMode20261022220100  = "repeat_mode"
	repeats20261022220100     = "repeats"

	repeatFilterBatchSize20261022220100 = 500
)

type repeatFilterStats20261022220100 struct {
	filters      int // filter strings with at least one translated term
	sorts        int // sort_by entries translated
	untranslated int // repeat terms and sort entries left as they were
}

type filterTranslation20261022220100 struct {
	filter       string
	translated   int
	untranslated []string
}

// The filter strings are stored in the human syntax (like / in / not in, bare
// values). To read them exactly as the model does, the next three functions are
// frozen copies of pkg/models/task_collection_filter.go's quotedRunEnd,
// replaceFilterOperators and preprocessFilterString; migrations must not follow
// later model changes.

var filterOperatorSigils20261022220100 = []struct {
	operator string
	sigil    string
}{
	{" not in ", " " + string(fexpr.SignAnyNeq) + " "},
	{" in ", " " + string(fexpr.SignAnyEq) + " "},
	{" like ", " " + string(fexpr.SignLike) + " "},
}

func quotedRunEnd20261022220100(filter string, start int) int {
	quote := filter[start]
	for i := start + 1; i < len(filter); i++ {
		switch filter[i] {
		case '\\':
			i++
		case quote:
			return i + 1
		}
	}
	return -1
}

// replaceFilterOperators20261022220100 rewrites the human operators to fexpr sigils,
// skipping quoted values. It also returns origin: origin[i] is the offset in filter
// of byte i of the result, and origin[len(result)] is len(filter).
func replaceFilterOperators20261022220100(filter string) (string, []int) {
	var out strings.Builder
	out.Grow(len(filter))
	origin := make([]int, 0, len(filter)+1)

	for i := 0; i < len(filter); {
		if c := filter[i]; c == '\'' || c == '"' {
			if end := quotedRunEnd20261022220100(filter, i); end > 0 {
				out.WriteString(filter[i:end])
				for j := i; j < end; j++ {
					origin = append(origin, j)
				}
				i = end
				continue
			}
		}

		matched := false
		for _, op := range filterOperatorSigils20261022220100 {
			if strings.HasPrefix(filter[i:], op.operator) {
				out.WriteString(op.sigil)
				for j := range len(op.sigil) {
					origin = append(origin, i+min(j, len(op.operator)-1))
				}
				i += len(op.operator)
				matched = true
				break
			}
		}
		if !matched {
			out.WriteByte(filter[i])
			origin = append(origin, i)
			i++
		}
	}

	return out.String(), append(origin, len(filter))
}

var filterTermPattern20261022220100 = regexp.MustCompile(`(\w+)\s*(>=|<=|!=|~|\?=|\?!=|=|>|<)\s*([^&|()]+)`)

func isQuoted20261022220100(value string) bool {
	return (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) ||
		(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""))
}

// quoteFilterValues20261022220100 quotes bare values so fexpr.Parse accepts them.
// It runs on the output of replaceFilterOperators20261022220100.
func quoteFilterValues20261022220100(filter string) string {
	return filterTermPattern20261022220100.ReplaceAllStringFunc(filter, func(match string) string {
		parts := filterTermPattern20261022220100.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		field := parts[1]
		comparator := parts[2]
		value := strings.TrimSpace(parts[3])

		if isQuoted20261022220100(value) {
			return field + " " + comparator + " " + value
		}

		quotedValue := "'" + strings.ReplaceAll(value, "'", "\\'") + "'"
		return field + " " + comparator + " " + quotedValue
	})
}

func parseFilter20261022220100(filter string) ([]fexpr.ExprGroup, error) {
	ops, _ := replaceFilterOperators20261022220100(filter)
	return fexpr.Parse(quoteFilterValues20261022220100(ops))
}

func isRepeatField20261022220100(name string) bool {
	return name == repeatAfter20261022220100 || name == repeatMode20261022220100
}

func isZero20261022220100(literal string) bool {
	n, err := strconv.ParseInt(strings.TrimSpace(literal), 10, 64)
	return err == nil && n == 0
}

// isRepeatTerm20261022220100 reports whether an expression compares a legacy repeat
// field: `repeat_after > 0`, or the mirrored `0 < repeat_after`.
func isRepeatTerm20261022220100(e fexpr.Expr) bool {
	if e.Left.Type == fexpr.TokenIdentifier && isRepeatField20261022220100(e.Left.Literal) {
		return true
	}
	return e.Left.Type == fexpr.TokenNumber && isRepeatField20261022220100(e.Right.Literal)
}

// repeatReplacement20261022220100 returns the repeats predicate equivalent to e, or
// false when there is none. Only the existence forms of repeat_after translate:
// `> 0` and `!= 0` mean "repeats", `= 0` means "doesn't repeat". 20261022220000 normalises
// repeats NULL to the empty string, so comparing with the empty string is exact.
func repeatReplacement20261022220100(e fexpr.Expr) (fexpr.SignOp, bool) {
	var op fexpr.SignOp
	switch {
	case e.Left.Type == fexpr.TokenIdentifier && e.Left.Literal == repeatAfter20261022220100 &&
		(e.Right.Type == fexpr.TokenText || e.Right.Type == fexpr.TokenNumber) && isZero20261022220100(e.Right.Literal):
		op = e.Op
	case e.Left.Type == fexpr.TokenNumber && isZero20261022220100(e.Left.Literal) && e.Right.Literal == repeatAfter20261022220100:
		// Mirror `0 < repeat_after` into `repeat_after > 0`.
		switch e.Op {
		case fexpr.SignLt:
			op = fexpr.SignGt
		case fexpr.SignEq, fexpr.SignNeq:
			op = e.Op
		}
	}

	switch op {
	case fexpr.SignGt, fexpr.SignNeq:
		return fexpr.SignNeq, true
	case fexpr.SignEq:
		return fexpr.SignEq, true
	}
	return "", false
}

func repeatsExpr20261022220100(op fexpr.SignOp) fexpr.Expr {
	return fexpr.Expr{
		Left:  fexpr.Token{Type: fexpr.TokenIdentifier, Literal: repeats20261022220100},
		Op:    op,
		Right: fexpr.Token{Type: fexpr.TokenText, Literal: ""},
	}
}

var humanOperators20261022220100 = map[fexpr.SignOp]string{
	fexpr.SignLike:   "like",
	fexpr.SignAnyEq:  "in",
	fexpr.SignAnyNeq: "not in",
}

func describeTerm20261022220100(e fexpr.Expr) string {
	op, ok := humanOperators20261022220100[e.Op]
	if !ok {
		op = string(e.Op)
	}
	return e.Left.Literal + " " + op + " " + e.Right.Literal
}

// walkRepeatTerms20261022220100 visits the repeat terms of a filter tree in text
// order and returns a copy of the tree in which each term is replaced by what visit
// returns.
func walkRepeatTerms20261022220100(groups []fexpr.ExprGroup, visit func(fexpr.Expr) fexpr.Expr) []fexpr.ExprGroup {
	out := make([]fexpr.ExprGroup, 0, len(groups))
	for _, g := range groups {
		switch item := g.Item.(type) {
		case fexpr.Expr:
			if isRepeatTerm20261022220100(item) {
				g.Item = visit(item)
			}
		case []fexpr.ExprGroup:
			g.Item = walkRepeatTerms20261022220100(item, visit)
		}
		out = append(out, g)
	}
	return out
}

type filterSpan20261022220100 struct{ start, end int }

// locateRepeatTerms20261022220100 returns the byte span in filter of every repeat
// term, in text order. It finds the terms the way the model's preprocessing does,
// so the n-th span is the n-th repeat term of the parsed tree.
func locateRepeatTerms20261022220100(filter string) []filterSpan20261022220100 {
	ops, origin := replaceFilterOperators20261022220100(filter)
	var spans []filterSpan20261022220100
	for _, m := range filterTermPattern20261022220100.FindAllStringSubmatchIndex(ops, -1) {
		field := ops[m[2]:m[3]]
		value := strings.TrimSpace(ops[m[6]:m[7]])
		if isQuoted20261022220100(value) && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		_, numErr := strconv.ParseInt(field, 10, 64)
		if !isRepeatField20261022220100(field) && (numErr != nil || !isRepeatField20261022220100(value)) {
			continue
		}
		valueEnd := m[6] + len(strings.TrimRight(ops[m[6]:m[7]], " \t\r\n"))
		spans = append(spans, filterSpan20261022220100{start: origin[m[2]], end: origin[valueEnd]})
	}
	return spans
}

// quotedRanges20261022220100 returns the quoted runs of filter, by the model's quoting rules.
func quotedRanges20261022220100(filter string) []filterSpan20261022220100 {
	var ranges []filterSpan20261022220100
	for i := 0; i < len(filter); i++ {
		if c := filter[i]; c == '\'' || c == '"' {
			if end := quotedRunEnd20261022220100(filter, i); end > 0 {
				ranges = append(ranges, filterSpan20261022220100{start: i, end: end})
				i = end - 1
			}
		}
	}
	return ranges
}

// insideQuotes20261022220100 reports whether a term starts inside a quoted run, or
// ends inside one that it started outside of. A term may contain a whole quoted
// value (`repeat_after = '0'`).
func insideQuotes20261022220100(span filterSpan20261022220100, ranges []filterSpan20261022220100) bool {
	for _, r := range ranges {
		startsInside := r.start <= span.start && span.start < r.end
		endsInside := span.start < r.start && r.start < span.end && span.end < r.end
		if startsInside || endsInside {
			return true
		}
	}
	return false
}

// translateRepeatFilter20261022220100 translates the existence forms of repeat_after
// in a stored filter string to repeats.
//
// The filter is parsed with fexpr after the model's own preprocessing, which decides
// what is a term. The rewrite then replaces only the byte span of each translated
// term, so the rest of the user's text (spacing, quoting, like / in) stays as it
// was; printing the tree back would have rewritten it in fexpr's syntax. The result
// must parse to the original tree with exactly those terms replaced, or the filter
// is left as it was. A filter without repeat terms comes back byte-identical.
func translateRepeatFilter20261022220100(filter string) (tr filterTranslation20261022220100) {
	tr.filter = filter
	if !strings.Contains(filter, repeatAfter20261022220100) && !strings.Contains(filter, repeatMode20261022220100) {
		return tr
	}

	tree, err := parseFilter20261022220100(filter)
	if err != nil {
		tr.untranslated = []string{fmt.Sprintf("%q (the filter does not parse: %v)", filter, err)}
		return tr
	}

	var terms []fexpr.Expr
	walkRepeatTerms20261022220100(tree, func(e fexpr.Expr) fexpr.Expr {
		terms = append(terms, e)
		return e
	})
	if len(terms) == 0 {
		return tr
	}

	describeAll := func(reason string) []string {
		out := make([]string, 0, len(terms))
		for _, e := range terms {
			out = append(out, describeTerm20261022220100(e)+reason)
		}
		return out
	}

	spans := locateRepeatTerms20261022220100(filter)
	if len(spans) != len(terms) {
		tr.untranslated = describeAll(" (could not locate the term in the filter text)")
		return tr
	}

	quoted := quotedRanges20261022220100(filter)
	translate := make([]bool, len(terms))
	var out strings.Builder
	last := 0
	for i, e := range terms {
		op, ok := repeatReplacement20261022220100(e)
		if !ok || insideQuotes20261022220100(spans[i], quoted) {
			tr.untranslated = append(tr.untranslated, describeTerm20261022220100(e))
			continue
		}
		translate[i] = true
		tr.translated++
		out.WriteString(filter[last:spans[i].start])
		out.WriteString(repeats20261022220100 + " " + string(op) + " ''")
		last = spans[i].end
	}
	if tr.translated == 0 {
		return tr
	}
	out.WriteString(filter[last:])

	if !rewriteVerifies20261022220100(tree, translate, out.String()) {
		tr.translated = 0
		tr.untranslated = describeAll(" (the rewritten filter did not verify)")
		return tr
	}

	tr.filter = out.String()
	return tr
}

// rewriteVerifies20261022220100 reports whether rewritten parses to tree with exactly
// the repeat terms marked in translate replaced by their repeats predicate.
func rewriteVerifies20261022220100(tree []fexpr.ExprGroup, translate []bool, rewritten string) bool {
	i := 0
	want := walkRepeatTerms20261022220100(tree, func(e fexpr.Expr) fexpr.Expr {
		defer func() { i++ }()
		if i >= len(translate) || !translate[i] {
			return e
		}
		op, _ := repeatReplacement20261022220100(e)
		return repeatsExpr20261022220100(op)
	})
	got, err := parseFilter20261022220100(rewritten)
	return err == nil && reflect.DeepEqual(got, want)
}

func marshalJSON20261022220100(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// Filters are full of && and <; keep them readable.
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func isEmptyJSON20261022220100(raw []byte) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func mentionsRepeat20261022220100(raw []byte) bool {
	return bytes.Contains(raw, []byte(repeatAfter20261022220100)) || bytes.Contains(raw, []byte(repeatMode20261022220100))
}

// translateFilterString20261022220100 translates one filter string and logs what it
// leaves.
func translateFilterString20261022220100(filter, record string, stats *repeatFilterStats20261022220100) (string, bool) {
	tr := translateRepeatFilter20261022220100(filter)
	for _, term := range tr.untranslated {
		log.Warningf("Repeat filter migration: %s: left %s untranslated", record, term)
		stats.untranslated++
	}
	if tr.translated == 0 {
		return filter, false
	}
	stats.filters++
	return tr.filter, true
}

// translateTaskCollection20261022220100 translates one stored TaskCollection JSON
// object: its filter string and its sort_by entries. order_by is index-aligned with
// sort_by and stays as it is. Unknown keys are kept. It returns raw unchanged unless
// something was translated.
func translateTaskCollection20261022220100(raw []byte, record string, stats *repeatFilterStats20261022220100) ([]byte, bool) {
	if isEmptyJSON20261022220100(raw) || !mentionsRepeat20261022220100(raw) {
		return raw, false
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		log.Warningf("Repeat filter migration: %s: left untranslated, it is not a filter object: %s", record, raw)
		stats.untranslated++
		return raw, false
	}

	changed := false
	var filter string
	if f, ok := fields["filter"]; ok && json.Unmarshal(f, &filter) == nil {
		if translated, ok := translateFilterString20261022220100(filter, record, stats); ok {
			encoded, err := marshalJSON20261022220100(translated)
			if err != nil {
				log.Warningf("Repeat filter migration: %s: left untranslated, could not encode the filter: %v", record, err)
				return raw, false
			}
			fields["filter"] = encoded
			changed = true
		}
	}

	var sortBy []string
	if sb, ok := fields["sort_by"]; ok && json.Unmarshal(sb, &sortBy) == nil {
		sorts := 0
		for i, field := range sortBy {
			switch field {
			case repeatAfter20261022220100:
				sortBy[i] = repeats20261022220100
				sorts++
			case repeatMode20261022220100:
				log.Warningf("Repeat filter migration: %s: left sort_by %s untranslated, it has no equivalent", record, field)
				stats.untranslated++
			}
		}
		if sorts > 0 {
			encoded, err := marshalJSON20261022220100(sortBy)
			if err != nil {
				log.Warningf("Repeat filter migration: %s: left untranslated, could not encode sort_by: %v", record, err)
				return raw, false
			}
			fields["sort_by"] = encoded
			stats.sorts += sorts
			changed = true
		}
	}

	if !changed {
		return raw, false
	}
	out, err := marshalJSON20261022220100(fields)
	if err != nil {
		log.Warningf("Repeat filter migration: %s: left untranslated, could not encode it: %v", record, err)
		return raw, false
	}
	return out, true
}

// translateBucketConfiguration20261022220100 translates the filter of every bucket of
// a filter-mode kanban view. A bucket filter is a TaskCollection object, or a bare
// filter string in the format before 20241119115012.
func translateBucketConfiguration20261022220100(raw []byte, record string, stats *repeatFilterStats20261022220100) ([]byte, bool) {
	if isEmptyJSON20261022220100(raw) || !mentionsRepeat20261022220100(raw) {
		return raw, false
	}

	var buckets []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &buckets); err != nil {
		log.Warningf("Repeat filter migration: %s: left untranslated, the bucket configuration is not a list: %s", record, raw)
		stats.untranslated++
		return raw, false
	}

	changed := false
	for i, bucket := range buckets {
		filter, ok := bucket["filter"]
		if !ok {
			continue
		}
		bucketRecord := fmt.Sprintf("%s bucket %d", record, i)

		var filterString string
		if json.Unmarshal(filter, &filterString) == nil {
			translated, ok := translateFilterString20261022220100(filterString, bucketRecord, stats)
			if !ok {
				continue
			}
			encoded, err := marshalJSON20261022220100(translated)
			if err != nil {
				log.Warningf("Repeat filter migration: %s: left untranslated, could not encode the filter: %v", bucketRecord, err)
				continue
			}
			bucket["filter"] = encoded
			changed = true
			continue
		}

		if translated, ok := translateTaskCollection20261022220100(filter, bucketRecord, stats); ok {
			bucket["filter"] = translated
			changed = true
		}
	}

	if !changed {
		return raw, false
	}
	out, err := marshalJSON20261022220100(buckets)
	if err != nil {
		log.Warningf("Repeat filter migration: %s: left untranslated, could not encode the bucket configuration: %v", record, err)
		return raw, false
	}
	return out, true
}

func migrateSavedFilters20261022220100(tx *xorm.Engine, stats *repeatFilterStats20261022220100) error {
	var lastID int64
	for {
		var page []savedFilter20261022220100
		err := tx.Cols("id", "filters").
			Where("id > ?", lastID).
			OrderBy("id ASC").
			Limit(repeatFilterBatchSize20261022220100).
			Find(&page)
		if err != nil {
			return fmt.Errorf("could not read saved filters after %d: %w", lastID, err)
		}
		if len(page) == 0 {
			return nil
		}

		for _, sf := range page {
			lastID = sf.ID
			out, changed := translateTaskCollection20261022220100([]byte(sf.Filters), fmt.Sprintf("saved filter %d", sf.ID), stats)
			if !changed {
				continue
			}
			if _, err := tx.Exec("UPDATE saved_filters SET filters = ? WHERE id = ?", string(out), sf.ID); err != nil {
				return fmt.Errorf("could not update saved filter %d: %w", sf.ID, err)
			}
		}
	}
}

func migrateProjectViews20261022220100(tx *xorm.Engine, stats *repeatFilterStats20261022220100) error {
	var lastID int64
	for {
		var page []projectView20261022220100
		err := tx.Cols("id", "filter", "bucket_configuration").
			Where("id > ?", lastID).
			OrderBy("id ASC").
			Limit(repeatFilterBatchSize20261022220100).
			Find(&page)
		if err != nil {
			return fmt.Errorf("could not read project views after %d: %w", lastID, err)
		}
		if len(page) == 0 {
			return nil
		}

		for _, view := range page {
			lastID = view.ID
			record := fmt.Sprintf("project view %d", view.ID)

			if out, changed := translateTaskCollection20261022220100([]byte(view.Filter), record, stats); changed {
				if _, err := tx.Exec("UPDATE project_views SET filter = ? WHERE id = ?", string(out), view.ID); err != nil {
					return fmt.Errorf("could not update the filter of project view %d: %w", view.ID, err)
				}
			}
			if out, changed := translateBucketConfiguration20261022220100([]byte(view.BucketConfiguration), record, stats); changed {
				if _, err := tx.Exec("UPDATE project_views SET bucket_configuration = ? WHERE id = ?", string(out), view.ID); err != nil {
					return fmt.Errorf("could not update the bucket configuration of project view %d: %w", view.ID, err)
				}
			}
		}
	}
}

// migrateRepeatFilters20261022220100 translates saved filters, view filters and
// bucket filters from the legacy repeat_after field to repeats. Records are written
// back only when something was translated, so a re-run changes nothing.
func migrateRepeatFilters20261022220100(tx *xorm.Engine) error {
	stats := &repeatFilterStats20261022220100{}
	if err := migrateSavedFilters20261022220100(tx, stats); err != nil {
		return err
	}
	if err := migrateProjectViews20261022220100(tx, stats); err != nil {
		return err
	}
	log.Infof("Repeat filter migration: %d filters translated, %d sorts translated, %d terms left untranslated", stats.filters, stats.sorts, stats.untranslated)
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261022220100",
		Description: "Translate saved filters and views from the legacy repeat_after field to repeats",
		Migrate:     migrateRepeatFilters20261022220100,
		Rollback: func(tx *xorm.Engine) error {
			// Nothing to undo: repeat_after no longer exists after 20261022220000.
			return nil
		},
	})
}
