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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/ganigeorgiev/fexpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// savedFiltersBefore20261022220100 is the saved_filters table as upstream creates it
// (pkg/models/saved_filters.go). It is only used for the DDL: rows are inserted with
// raw SQL so the stored JSON text is exactly what the test wrote.
type savedFiltersBefore20261022220100 struct {
	ID          int64     `xorm:"autoincr not null unique pk"`
	Filters     string    `xorm:"JSON not null"`
	Title       string    `xorm:"varchar(250) not null"`
	Description string    `xorm:"longtext null"`
	OwnerID     int64     `xorm:"bigint not null INDEX"`
	IsFavorite  bool      `xorm:"default false"`
	Created     time.Time `xorm:"created not null"`
	Updated     time.Time `xorm:"updated not null"`
}

func (savedFiltersBefore20261022220100) TableName() string { return "saved_filters" }

// projectViewsBefore20261022220100 is the project_views table as upstream creates it
// (pkg/models/project_view.go), with JSON filter columns. DDL only, like above.
type projectViewsBefore20261022220100 struct {
	ID                      int64     `xorm:"autoincr not null unique pk"`
	Title                   string    `xorm:"varchar(255) not null"`
	ProjectID               int64     `xorm:"not null index"`
	ViewKind                int       `xorm:"not null"`
	Filter                  string    `xorm:"JSON null default null"`
	Position                float64   `xorm:"double null"`
	BucketConfigurationMode int       `xorm:"default 0"`
	BucketConfiguration     string    `xorm:"JSON"`
	DefaultBucketID         int64     `xorm:"bigint INDEX null"`
	DoneBucketID            int64     `xorm:"bigint INDEX null"`
	Updated                 time.Time `xorm:"updated not null"`
	Created                 time.Time `xorm:"created not null"`
}

func (projectViewsBefore20261022220100) TableName() string { return "project_views" }

func captureLog20261022220100(t *testing.T) func() string {
	t.Helper()
	dir := t.TempDir()
	log.ConfigureStandardLogger(true, "file", dir, "INFO", "text")
	t.Cleanup(log.InitLogger)
	return func() string {
		b, err := os.ReadFile(filepath.Join(dir, "standard.log"))
		require.NoError(t, err)
		return string(b)
	}
}

func dropFilterTables20261022220100(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.DropTables("saved_filters", "project_views"))
}

func filterTestEngine20261022220100(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	dropFilterTables20261022220100(t, x)
	t.Cleanup(func() { dropFilterTables20261022220100(t, x) })
	require.NoError(t, x.Sync2(savedFiltersBefore20261022220100{}, projectViewsBefore20261022220100{}))
	return x
}

var seedTime20261022220100 = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

func insertSavedFilter20261022220100(t *testing.T, x *xorm.Engine, id int64, filters string) {
	t.Helper()
	_, err := x.Exec("INSERT INTO saved_filters (id, filters, title, owner_id, is_favorite, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, filters, fmt.Sprintf("filter %d", id), 1, false, seedTime20261022220100, seedTime20261022220100)
	require.NoError(t, err)
}

// insertProjectView20261022220100 stores a view; nil stores NULL.
func insertProjectView20261022220100(t *testing.T, x *xorm.Engine, id int64, filter, buckets *string) {
	t.Helper()
	var f, b any
	if filter != nil {
		f = *filter
	}
	if buckets != nil {
		b = *buckets
	}
	_, err := x.Exec("INSERT INTO project_views (id, title, project_id, view_kind, filter, position, bucket_configuration_mode, bucket_configuration, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, fmt.Sprintf("view %d", id), 1, 3, f, 100, 2, b, seedTime20261022220100, seedTime20261022220100)
	require.NoError(t, err)
}

func strptr20261022220100(s string) *string { return &s }

// filterRows20261022220100 reads every stored filter as raw text, keyed by table and id.
func filterRows20261022220100(t *testing.T, x *xorm.Engine) map[string]string {
	t.Helper()
	rows := map[string]string{}
	sf, err := x.QueryString("SELECT id, filters FROM saved_filters ORDER BY id")
	require.NoError(t, err)
	for _, r := range sf {
		rows["saved_filters "+r["id"]] = r["filters"]
	}
	pv, err := x.QueryString("SELECT id, filter, bucket_configuration FROM project_views ORDER BY id")
	require.NoError(t, err)
	for _, r := range pv {
		rows["project_views.filter "+r["id"]] = r["filter"]
		rows["project_views.bucket_configuration "+r["id"]] = r["bucket_configuration"]
	}
	return rows
}

// assertStoredJSON20261022220100 checks a stored value byte for byte: the migration
// writes compact JSON with sorted keys and without HTML escaping.
func assertStoredJSON20261022220100(t *testing.T, want, got string, msgAndArgs ...any) bool {
	t.Helper()
	return assert.JSONEq(t, want, got, msgAndArgs...) && assert.Equal(t, want, got, msgAndArgs...)
}

func TestRepeatFilterMigrationRegistered20261022220100(t *testing.T) {
	var ids []string
	for _, m := range migrations {
		ids = append(ids, m.ID)
	}
	sort.Strings(ids)
	require.Contains(t, ids, "20261022220000")
	require.Contains(t, ids, "20261022220100")
	// It runs right after the RRULE conversion, which normalises repeats NULL to ''.
	for i, id := range ids {
		if id == "20261022220000" {
			require.Less(t, i+1, len(ids))
			assert.Equal(t, "20261022220100", ids[i+1], "nothing may run between the RRULE conversion and the filter translation")
		}
	}
}

func TestTranslateRepeatFilter20261022220100(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		want         string // "" means unchanged
		translated   int
		untranslated int
	}{
		{name: "greater than zero", in: "repeat_after > 0", want: "repeats != ''", translated: 1},
		{name: "not zero", in: "repeat_after != 0", want: "repeats != ''", translated: 1},
		{name: "equals zero", in: "repeat_after = 0", want: "repeats = ''", translated: 1},
		{name: "no spaces", in: "repeat_after>0", want: "repeats != ''", translated: 1},
		{name: "quoted zero", in: "repeat_after = '0'", want: "repeats = ''", translated: 1},
		{name: "mirrored less than", in: "0 < repeat_after", want: "repeats != ''", translated: 1},
		{name: "mirrored not equal", in: "0 != repeat_after", want: "repeats != ''", translated: 1},
		{name: "mirrored equals", in: "0 = repeat_after", want: "repeats = ''", translated: 1},
		{name: "joined", in: "done = false && repeat_after > 0", want: "done = false && repeats != ''", translated: 1},
		{name: "group", in: "(repeat_after = 0 || priority > 2)", want: "(repeats = '' || priority > 2)", translated: 1},
		{
			name:       "nested group",
			in:         "done = true && (priority > 2 || (repeat_after > 0 && labels in 1, 2))",
			want:       "done = true && (priority > 2 || (repeats != '' && labels in 1, 2))",
			translated: 1,
		},
		{
			name:       "after a rewritten operator",
			in:         "labels in 1, 2 || title like '%x%' || repeat_after != 0",
			want:       "labels in 1, 2 || title like '%x%' || repeats != ''",
			translated: 1,
		},
		{
			name:       "quoted text before the term",
			in:         "title like '%repeat_after > 0%' && repeat_after = 0",
			want:       "title like '%repeat_after > 0%' && repeats = ''",
			translated: 1,
		},
		{
			name:       "two terms, one translatable",
			in:         "repeat_after > 0 && repeat_after < 86400",
			want:       "repeats != '' && repeat_after < 86400",
			translated: 1, untranslated: 1,
		},
		{name: "two translatable terms", in: "repeat_after = 0 || repeat_after > 0", want: "repeats = '' || repeats != ''", translated: 2},
		{name: "date math kept", in: "due_date < now+7d && repeat_after > 0", want: "due_date < now+7d && repeats != ''", translated: 1},

		{name: "other value", in: "repeat_after > 86400", untranslated: 1},
		{name: "other operator", in: "repeat_after >= 0", untranslated: 1},
		{name: "in list", in: "repeat_after in 0, 86400", untranslated: 1},
		{name: "repeat mode", in: "repeat_mode = 1", untranslated: 1},
		{name: "mirrored other value", in: "5 < repeat_after", untranslated: 1},
		{name: "unparsable", in: "repeat_after > 0 &&", untranslated: 1},
		// The model's preprocessing ends a value at & even inside quotes, so it cannot parse this either.
		{name: "unparsable by the model", in: "title like 'a & b' && repeat_after > 0", untranslated: 1},
		// The model's preprocessing splits these quoted values at && and sees a term
		// inside them; text inside quotes is never rewritten. The first does not
		// parse; the second parses, and would verify if it were rewritten.
		{name: "term inside an unparsable quoted value", in: "title = 'x && repeat_after > 0 && y'", untranslated: 1},
		{name: "term inside a quoted value", in: "title = 'x && repeat_after > 0 && y = z'", untranslated: 1},

		{name: "quoted pattern", in: "title like '%repeat_after > 0%'"},
		{name: "double-quoted pattern", in: `description = "repeat_after = 0"`},
		{name: "value named like the field", in: "title = repeat_after"},
		{name: "no repeat terms", in: "done = false && priority >= 3"},
		{name: "empty", in: ""},
		{name: "already translated", in: "done = false && repeats != ''"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := translateRepeatFilter20261022220100(c.in)
			want := c.want
			if want == "" {
				want = c.in
			}
			assert.Equal(t, want, got.filter)
			assert.Equal(t, c.translated, got.translated, "translated terms")
			assert.Len(t, got.untranslated, c.untranslated, "untranslated terms: %v", got.untranslated)

			// Idempotent: the output translates to itself.
			again := translateRepeatFilter20261022220100(got.filter)
			assert.Equal(t, got.filter, again.filter)
			assert.Equal(t, 0, again.translated)
		})
	}
}

// The rewrite changes exactly the translated terms: the output parses (with the
// model's preprocessing) to the input's tree with only those terms replaced.
func TestTranslateRepeatFilterKeepsTree20261022220100(t *testing.T) {
	in := "(title like 'a b' || labels not in 3, 4) && (repeat_after != 0 && (due_date >= now/d || done = true))"
	got := translateRepeatFilter20261022220100(in)
	require.Equal(t, 1, got.translated)

	parse := func(s string) []fexpr.ExprGroup {
		ops, _ := replaceFilterOperators20261022220100(s)
		tree, err := fexpr.Parse(quoteFilterValues20261022220100(ops))
		require.NoError(t, err, s)
		return tree
	}
	want := parse(strings.Replace(in, "repeat_after != 0", "repeats != ''", 1))
	assert.Equal(t, want, parse(got.filter))
	assert.NotEqual(t, parse(in), parse(got.filter))
}

func TestInsideQuotes20261022220100(t *testing.T) {
	filter := "title = 'x && repeat_after > 0' && repeat_after = '0'"
	quoted := quotedRanges20261022220100(filter)
	require.Len(t, quoted, 2)
	span := func(term string, from int) filterSpan20261022220100 {
		start := strings.Index(filter[from:], term) + from
		return filterSpan20261022220100{start: start, end: start + len(term)}
	}
	assert.True(t, insideQuotes20261022220100(span("repeat_after > 0", 0), quoted), "starts inside a quoted value")
	assert.False(t, insideQuotes20261022220100(span("repeat_after = '0'", 30), quoted), "contains a whole quoted value")
	assert.True(t, insideQuotes20261022220100(filterSpan20261022220100{start: 0, end: 10}, quoted), "ends inside a quoted value")
	assert.False(t, insideQuotes20261022220100(filterSpan20261022220100{start: 0, end: 5}, quoted), "before any quote")
}

func TestRewriteVerifies20261022220100(t *testing.T) {
	tree, err := parseFilter20261022220100("done = true && repeat_after > 0")
	require.NoError(t, err)
	assert.True(t, rewriteVerifies20261022220100(tree, []bool{true}, "done = true && repeats != ''"))
	assert.True(t, rewriteVerifies20261022220100(tree, []bool{false}, "done = true && repeat_after > 0"))
	assert.False(t, rewriteVerifies20261022220100(tree, []bool{true}, "done = true && repeats = ''"), "wrong predicate")
	assert.False(t, rewriteVerifies20261022220100(tree, []bool{true}, "done = false && repeats != ''"), "another term changed")
	assert.False(t, rewriteVerifies20261022220100(tree, []bool{true}, "done = true || repeats != ''"), "join changed")
	assert.False(t, rewriteVerifies20261022220100(tree, []bool{true}, "done = true && (repeats != ''"), "does not parse")
}

func TestTranslateTaskCollection20261022220100(t *testing.T) {
	t.Run("filter and sort, unknown keys kept", func(t *testing.T) {
		stats := &repeatFilterStats20261022220100{}
		in := `{"filter":"repeat_after > 0 && title like '<a>'","order_by":["desc","asc"],"sort_by":["repeat_after","id"],"x_future":{"nested":[1,"two"]}}`
		out, changed := translateTaskCollection20261022220100([]byte(in), "saved filter 1", stats)
		require.True(t, changed)
		assert.JSONEq(t, `{"filter":"repeats != '' && title like '<a>'","order_by":["desc","asc"],"sort_by":["repeats","id"],"x_future":{"nested":[1,"two"]}}`, string(out))
		assert.Contains(t, string(out), `"repeats != '' && title like '<a>'"`, "no HTML escaping")
		assert.Equal(t, repeatFilterStats20261022220100{filters: 1, sorts: 1}, *stats)
	})
	t.Run("sort only, order_by aligned", func(t *testing.T) {
		stats := &repeatFilterStats20261022220100{}
		in := `{"filter":"","order_by":["asc","desc","asc"],"sort_by":["due_date","repeat_after","repeat_mode"]}`
		out, changed := translateTaskCollection20261022220100([]byte(in), "saved filter 2", stats)
		require.True(t, changed)
		assertStoredJSON20261022220100(t, `{"filter":"","order_by":["asc","desc","asc"],"sort_by":["due_date","repeats","repeat_mode"]}`, string(out))
		assert.Equal(t, repeatFilterStats20261022220100{sorts: 1, untranslated: 1}, *stats)
	})
	for _, in := range []string{
		`{"filter":"done = true","sort_by":["priority"]}`,
		`{"filter":"repeat_after > 86400"}`,
		`{"sort_by":["repeat_mode"],"order_by":["asc"]}`,
		`null`,
		``,
		`not json at all repeat_after`,
		`{"filter":42}`,
	} {
		t.Run("unchanged "+in, func(t *testing.T) {
			stats := &repeatFilterStats20261022220100{}
			out, changed := translateTaskCollection20261022220100([]byte(in), "saved filter 3", stats)
			assert.False(t, changed)
			assert.Equal(t, in, string(out))
			assert.Zero(t, stats.filters)
			assert.Zero(t, stats.sorts)
		})
	}
}

func TestRepeatFilterMigration20261022220100(t *testing.T) {
	x := filterTestEngine20261022220100(t)
	readLog := captureLog20261022220100(t)

	untouched := map[int64]string{
		3: `{"filter":"repeat_after > 86400"}`,
		4: `{"filter":"title like '%repeat_after > 0%'"}`,
		5: `{"filter":"done = true","order_by":["desc"],"sort_by":["priority"]}`,
		6: `{"filter":"priority >= 3","order_by":["asc"],"sort_by":["repeat_mode"]}`,
	}
	insertSavedFilter20261022220100(t, x, 1, `{"filter":"done = false && repeat_after > 0","filter_include_nulls":false,"s":""}`)
	insertSavedFilter20261022220100(t, x, 2, `{"filter":"(repeat_after = 0 || priority > 2)","order_by":["asc","desc","asc"],"s":"","sort_by":["due_date","repeat_after","id"],"x_legacy":{"keep":[1,2]}}`)
	for id, f := range untouched {
		insertSavedFilter20261022220100(t, x, id, f)
	}
	// More than one page, so paging is exercised.
	const bulk = 505
	for i := int64(0); i < bulk; i++ {
		insertSavedFilter20261022220100(t, x, 100+i, `{"filter":"repeat_after > 0"}`)
	}

	insertProjectView20261022220100(t, x, 1, strptr20261022220100(`{"filter":"repeat_after != 0"}`), nil)
	insertProjectView20261022220100(t, x, 2, nil, strptr20261022220100(
		`[{"filter":{"filter":"repeat_after > 0","order_by":["asc"],"sort_by":["repeat_after"]},"title":"Repeating"},{"filter":{"filter":"done = true"},"title":"Done"}]`))
	insertProjectView20261022220100(t, x, 3, strptr20261022220100(`{"filter":"priority > 2"}`), strptr20261022220100(`[]`))
	insertProjectView20261022220100(t, x, 4, strptr20261022220100(`{"filter":"repeat_after = 0"}`), strptr20261022220100(`[{"filter":"repeat_after > 0","title":"Old"}]`))

	before := filterRows20261022220100(t, x)
	require.NoError(t, migrateRepeatFilters20261022220100(x))
	after := filterRows20261022220100(t, x)

	assertStoredJSON20261022220100(t, `{"filter":"done = false && repeats != ''","filter_include_nulls":false,"s":""}`, after["saved_filters 1"])
	assertStoredJSON20261022220100(t, `{"filter":"(repeats = '' || priority > 2)","order_by":["asc","desc","asc"],"s":"","sort_by":["due_date","repeats","id"],"x_legacy":{"keep":[1,2]}}`, after["saved_filters 2"])
	for i := int64(0); i < bulk; i++ {
		if !assertStoredJSON20261022220100(t, `{"filter":"repeats != ''"}`, after[fmt.Sprintf("saved_filters %d", 100+i)], "bulk filter %d", 100+i) {
			break
		}
	}
	for id := range untouched {
		key := fmt.Sprintf("saved_filters %d", id)
		assert.Equal(t, before[key], after[key], "%s must stay byte-identical", key)
	}
	assert.Contains(t, before["saved_filters 4"], "repeat_after > 0", "the quoted case must contain the pattern")

	assertStoredJSON20261022220100(t, `{"filter":"repeats != ''"}`, after["project_views.filter 1"])
	assert.Empty(t, after["project_views.bucket_configuration 1"])
	assert.Empty(t, after["project_views.filter 2"])
	assertStoredJSON20261022220100(t, `[{"filter":{"filter":"repeats != ''","order_by":["asc"],"sort_by":["repeats"]},"title":"Repeating"},{"filter":{"filter":"done = true"},"title":"Done"}]`, after["project_views.bucket_configuration 2"])
	assert.Equal(t, before["project_views.filter 3"], after["project_views.filter 3"])
	assert.Equal(t, before["project_views.bucket_configuration 3"], after["project_views.bucket_configuration 3"])
	assertStoredJSON20261022220100(t, `{"filter":"repeats = ''"}`, after["project_views.filter 4"])
	assertStoredJSON20261022220100(t, `[{"filter":"repeats != ''","title":"Old"}]`, after["project_views.bucket_configuration 4"])

	logged := readLog()
	assert.Contains(t, logged, "Repeat filter migration: 511 filters translated, 2 sorts translated, 2 terms left untranslated")
	assert.Regexp(t, `level=WARN.*saved filter 3: left repeat_after > 86400 untranslated`, logged)
	assert.Regexp(t, `level=WARN.*saved filter 6: left sort_by repeat_mode untranslated`, logged)

	// A second run changes nothing.
	require.NoError(t, migrateRepeatFilters20261022220100(x))
	assert.Equal(t, after, filterRows20261022220100(t, x))
	assert.Contains(t, readLog(), "Repeat filter migration: 0 filters translated, 0 sorts translated, 2 terms left untranslated")
}

// syncModelTables20261022220100 creates the model's tables for a test and drops the
// ones it created afterwards.
func syncModelTables20261022220100(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.DropTables("saved_filters", "project_views", "tasks"))
	existing := map[string]bool{}
	metas, err := x.DBMetas()
	require.NoError(t, err)
	for _, m := range metas {
		existing[m.Name] = true
	}
	tables := append(models.GetTables(), &user.User{})
	require.NoError(t, x.Sync2(tables...))
	t.Cleanup(func() {
		metas, err := x.DBMetas()
		require.NoError(t, err)
		var created []any
		for _, m := range metas {
			if !existing[m.Name] {
				created = append(created, m.Name)
			}
		}
		require.NoError(t, x.DropTables(created...))
	})
}

// The translated filters work through the model: a saved filter read and evaluated
// by TaskCollection returns the repeating and the non-repeating tasks, including a
// task whose repeats was NULL before the RRULE migration normalised it.
func TestRepeatFilterMigrationThroughModel20261022220100(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	syncModelTables20261022220100(t, x)

	u := &user.User{Username: "rrule-filter-20261022220100", Email: "rrule-filter@example.com", Issuer: "local"}
	// A pre-existing users table may still hold this user from an interrupted run.
	_, err = x.Where("username = ?", u.Username).Delete(&user.User{})
	require.NoError(t, err)
	_, err = x.Insert(u)
	require.NoError(t, err)
	t.Cleanup(func() {
		// users may predate this test and survive its table cleanup.
		_, err := x.ID(u.ID).Delete(&user.User{})
		require.NoError(t, err)
	})
	project := &models.Project{Title: "rrule filters", OwnerID: u.ID}
	_, err = x.Insert(project)
	require.NoError(t, err)
	_, err = x.Insert(&models.ProjectAncestor{AncestorID: project.ID, ProjectID: project.ID, Depth: 0})
	require.NoError(t, err)

	newTask := func(index int64, title, repeats string) int64 {
		task := &models.Task{Title: title, ProjectID: project.ID, Index: index, CreatedByID: u.ID, Repeats: repeats}
		_, err := x.Insert(task)
		require.NoError(t, err)
		return task.ID
	}
	repeating := newTask(1, "repeating", "FREQ=DAILY;INTERVAL=1")
	plain := newTask(2, "plain", "")
	wasNull := newTask(3, "was null", "")
	_, err = x.Exec("UPDATE tasks SET repeats = NULL WHERE id = ?", wasNull)
	require.NoError(t, err)
	nulls, err := x.Where("repeats IS NULL").Count(&models.Task{})
	require.NoError(t, err)
	require.Equal(t, int64(1), nulls)

	// The RRULE migration finds the repeats column and only normalises NULL to ''.
	require.NoError(t, migrateRRule20261022220000(x))

	insertSavedFilter := func(id int64, filters string) {
		_, err := x.Exec("INSERT INTO saved_filters (id, filters, title, owner_id, is_favorite, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?)",
			id, filters, fmt.Sprintf("filter %d", id), u.ID, false, seedTime20261022220100, seedTime20261022220100)
		require.NoError(t, err)
	}
	insertSavedFilter(1, `{"filter":"repeat_after > 0"}`)
	insertSavedFilter(2, `{"filter":"repeat_after = 0","order_by":["desc"],"sort_by":["repeat_after"]}`)
	insertSavedFilter(3, `{"filter":"done = false","order_by":["desc","asc"],"sort_by":["repeat_after","id"]}`)

	read := func(filterID int64) ([]int64, error) {
		s := db.NewSession()
		defer s.Close()
		tc := &models.TaskCollection{ProjectID: filterID*-1 - 1}
		result, _, _, err := tc.ReadAll(s, u, "", 1, 50)
		if err != nil {
			return nil, err
		}
		tasks, ok := result.([]*models.Task)
		require.True(t, ok, "unexpected result type %T", result)
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		return ids, nil
	}

	// Before the translation the legacy field no longer works.
	for _, id := range []int64{1, 2, 3} {
		_, err := read(id)
		require.Error(t, err, "saved filter %d", id)
	}

	require.NoError(t, migrateRepeatFilters20261022220100(x))

	ids, err := read(1)
	require.NoError(t, err)
	assert.Equal(t, []int64{repeating}, ids)

	ids, err = read(2)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{plain, wasNull}, ids)

	// Sorted by repeats descending: the repeating task first.
	ids, err = read(3)
	require.NoError(t, err)
	require.Len(t, ids, 3)
	assert.Equal(t, repeating, ids[0])
}
