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

package caldavtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelationsBasic(t *testing.T) {
	// RFC 5545 §3.8.4.5 (rfc5545.txt line 6391):
	// "This property is used to represent a relationship or reference
	//  between one calendar component and another."

	t.Run("Parent with RELTYPE=CHILD and child with RELTYPE=PARENT", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create parent (no relations)
		parent := NewVTodo("rel-parent-1", "Parent Task").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/rel-parent-1.ics", parent)
		require.Equal(t, 201, rec.Code)

		// Create child referencing parent
		child := NewVTodo("rel-child-1", "Child Task").
			RelatedToParent("rel-parent-1").
			Build()
		rec = caldavPUT(t, e, "/dav/projects/36/rel-child-1.ics", child)
		require.Equal(t, 201, rec.Code)

		// GET child — should have RELATED-TO;RELTYPE=PARENT
		rec = caldavGET(t, e, "/dav/projects/36/rel-child-1.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:rel-parent-1",
			"Child should have RELATED-TO pointing to parent")

		// GET parent — should have RELATED-TO;RELTYPE=CHILD (inverse)
		rec = caldavGET(t, e, "/dav/projects/36/rel-parent-1.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:rel-child-1",
			"Parent should have inverse RELATED-TO pointing to child")
	})

	t.Run("Grandchild chain: parent -> child -> grandchild", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create in order: parent, child, grandchild
		parent := NewVTodo("rel-gp-parent", "Grandparent").Build()
		caldavPUT(t, e, "/dav/projects/36/rel-gp-parent.ics", parent)

		child := NewVTodo("rel-gp-child", "Parent").
			RelatedToParent("rel-gp-parent").
			Build()
		caldavPUT(t, e, "/dav/projects/36/rel-gp-child.ics", child)

		grandchild := NewVTodo("rel-gp-grandchild", "Child").
			RelatedToParent("rel-gp-child").
			Build()
		caldavPUT(t, e, "/dav/projects/36/rel-gp-grandchild.ics", grandchild)

		// Verify middle node has both parent and child relations
		rec := caldavGET(t, e, "/dav/projects/36/rel-gp-child.ics")
		body := rec.Body.String()
		assert.Contains(t, body, "RELATED-TO;RELTYPE=PARENT:rel-gp-parent")
		assert.Contains(t, body, "RELATED-TO;RELTYPE=CHILD:rel-gp-grandchild")
	})
}

func TestRelationsReverseOrder(t *testing.T) {
	t.Run("Child arrives before parent (Tasks.org pattern)", func(t *testing.T) {
		// This is the most common real-world scenario:
		// Tasks.org sends child with RELATED-TO;RELTYPE=PARENT but the parent
		// has NO RELATED-TO at all. The child may arrive before the parent.

		e := setupTestEnv(t)

		// Step 1: Child arrives first
		child := NewVTodo("rev-child-first", "Child First").
			RelatedToParent("rev-parent-late").
			Build()
		rec := caldavPUT(t, e, "/dav/projects/36/rev-child-first.ics", child)
		require.Equal(t, 201, rec.Code)

		// Step 2: Parent arrives later (no RELATED-TO)
		parent := NewVTodo("rev-parent-late", "Parent Late").Build()
		rec = caldavPUT(t, e, "/dav/projects/36/rev-parent-late.ics", parent)
		require.Equal(t, 201, rec.Code)

		// Step 3: Verify parent has correct title (not DUMMY-UID)
		rec = caldavGET(t, e, "/dav/projects/36/rev-parent-late.ics")
		assert.Contains(t, rec.Body.String(), "SUMMARY:Parent Late",
			"Parent should have its real title, not DUMMY-UID")
		assert.NotContains(t, rec.Body.String(), "DUMMY",
			"DUMMY placeholder should be replaced")

		// Step 4: Verify child still has parent relation
		rec = caldavGET(t, e, "/dav/projects/36/rev-child-first.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:rev-parent-late",
			"Child should still have parent relation after parent arrives")
	})

	t.Run("Multiple children before parent", func(t *testing.T) {
		e := setupTestEnv(t)

		// Two children arrive before parent
		child1 := NewVTodo("rev-mc1", "Multi Child 1").
			RelatedToParent("rev-mparent").Build()
		caldavPUT(t, e, "/dav/projects/36/rev-mc1.ics", child1)

		child2 := NewVTodo("rev-mc2", "Multi Child 2").
			RelatedToParent("rev-mparent").Build()
		caldavPUT(t, e, "/dav/projects/36/rev-mc2.ics", child2)

		// Parent arrives
		parent := NewVTodo("rev-mparent", "Multi Parent").Build()
		caldavPUT(t, e, "/dav/projects/36/rev-mparent.ics", parent)

		// Verify parent shows both children
		rec := caldavGET(t, e, "/dav/projects/36/rev-mparent.ics")
		body := rec.Body.String()
		assert.Contains(t, body, "RELATED-TO;RELTYPE=CHILD:rev-mc1")
		assert.Contains(t, body, "RELATED-TO;RELTYPE=CHILD:rev-mc2")
	})
}

func TestRelationsCrossProject(t *testing.T) {
	t.Run("Parent in project 36, child in project 38", func(t *testing.T) {
		e := setupTestEnv(t)

		parent := NewVTodo("xp-parent", "Cross-Project Parent").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/xp-parent.ics", parent)
		require.Equal(t, 201, rec.Code)

		child := NewVTodo("xp-child", "Cross-Project Child").
			RelatedToParent("xp-parent").Build()
		rec = caldavPUT(t, e, "/dav/projects/38/xp-child.ics", child)
		require.Equal(t, 201, rec.Code)

		// Verify parent in project 36 knows about child
		rec = caldavGET(t, e, "/dav/projects/36/xp-parent.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:xp-child",
			"Parent should have cross-project child relation")

		// Verify child in project 38 knows about parent
		rec = caldavGET(t, e, "/dav/projects/38/xp-child.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:xp-parent",
			"Child should have cross-project parent relation")
	})

	t.Run("Pre-existing cross-project relations from fixtures", func(t *testing.T) {
		e := setupTestEnv(t)

		// Task 45 (project 36) and task 46 (project 38) have cross-project relations in fixtures
		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test-parent-task-another-list.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task-another-list")

		rec = caldavGET(t, e, "/dav/projects/38/uid-caldav-test-child-task-another-list.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task-another-list")
	})
}

func TestRelationsDeletion(t *testing.T) {
	t.Run("Deleting child removes relation from parent", func(t *testing.T) {
		e := setupTestEnv(t)

		// Task 41 is parent of task 43 (from fixtures)
		rec := caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test-child-task.ics")
		assert.Equal(t, 204, rec.Code)

		// Parent should no longer reference deleted child
		rec = caldavGET(t, e, "/dav/projects/36/uid-caldav-test-parent-task.ics")
		assert.NotContains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task\r\n",
			"Parent should not reference deleted child")
	})

	t.Run("Deleting parent removes relation from child", func(t *testing.T) {
		e := setupTestEnv(t)

		// Delete parent task 41
		rec := caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test-parent-task.ics")
		assert.Equal(t, 204, rec.Code)

		// Child should no longer reference deleted parent
		rec = caldavGET(t, e, "/dav/projects/36/uid-caldav-test-child-task.ics")
		assert.NotContains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task",
			"Child should not reference deleted parent")
	})
}

func TestRelationsResync(t *testing.T) {
	t.Run("Parent re-sync without RELATED-TO preserves child relations", func(t *testing.T) {
		// This is the DAVx5 behavior: parent is updated (e.g., title change)
		// and re-synced without any RELATED-TO. The child-declared relations
		// should survive.

		e := setupTestEnv(t)

		// Create parent
		parent := NewVTodo("resync-parent", "Original Parent").Build()
		caldavPUT(t, e, "/dav/projects/36/resync-parent.ics", parent)

		// Create child with parent relation
		child := NewVTodo("resync-child", "Child").
			RelatedToParent("resync-parent").Build()
		caldavPUT(t, e, "/dav/projects/36/resync-child.ics", child)

		// Re-sync parent with updated title but NO RELATED-TO
		parentUpdated := NewVTodo("resync-parent", "Updated Parent Title").Build()
		caldavPUT(t, e, "/dav/projects/36/resync-parent.ics", parentUpdated)

		// Verify relations survived
		rec := caldavGET(t, e, "/dav/projects/36/resync-parent.ics")
		body := rec.Body.String()
		assert.Contains(t, body, "Updated Parent Title", "Title should be updated")
		assert.Contains(t, body, "RELATED-TO;RELTYPE=CHILD:resync-child",
			"Child relation should survive parent re-sync without RELATED-TO")

		rec = caldavGET(t, e, "/dav/projects/36/resync-child.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:resync-parent",
			"Parent relation on child should survive parent re-sync")
	})
}
