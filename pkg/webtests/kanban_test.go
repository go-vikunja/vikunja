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

package webtests

import (
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucket(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "GET", "/api/v1/projects/1/views/4/buckets", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `testbucket1`)
			assert.Contains(t, rec.Body.String(), `testbucket2`)
			assert.Contains(t, rec.Body.String(), `testbucket3`)
			assert.NotContains(t, rec.Body.String(), `testbucket4`) // Different Project
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the project was loaded successfully afterwards, see testReadOneWithUser
			rec, err := th.Request(t, "POST", "/api/v1/projects/1/views/4/buckets/1", strings.NewReader(`{"title":"TestLoremIpsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
		})
		t.Run("Nonexisting Bucket", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/1/views/4/buckets/9999", strings.NewReader(`{"title":"TestLoremIpsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/1/views/4/buckets/1", strings.NewReader(`{"title":""}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":400`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "POST", "/api/v1/projects/20/views/80/buckets/5", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/6/views/24/buckets/6", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/7/views/28/buckets/7", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/8/views/32/buckets/8", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/9/views/36/buckets/9", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/10/views/40/buckets/10", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/11/views/44/buckets/11", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/12/views/48/buckets/12", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/13/views/52/buckets/13", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/14/views/56/buckets/14", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/15/views/60/buckets/15", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/16/views/64/buckets/16", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/17/views/68/buckets/17", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "DELETE", "/api/v1/projects/1/views/4/buckets/1", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "DELETE", "/api/v1/projects/1/views/4/buckets/999", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "DELETE", "/api/v1/projects/20/views/80/buckets/5", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/6/views/24/buckets/6", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/7/views/28/buckets/7", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/8/views/32/buckets/8", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/9/views/36/buckets/9", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/10/views/40/buckets/10", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/11/views/44/buckets/11", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/12/views/48/buckets/12", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/13/views/52/buckets/13", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/14/views/56/buckets/14", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/15/views/60/buckets/15", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/16/views/64/buckets/16", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/17/views/68/buckets/17", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "PUT", "/api/v1/projects/1/views/3/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
		})
		t.Run("Nonexistent project", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/9999/views/1/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Nonexistent view", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/1/views/9999/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "PUT", "/api/v1/projects/20/views/80/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/6/views/24/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/7/views/28/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/8/views/32/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/9/views/36/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/10/views/40/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/11/views/44/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/12/views/48/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/13/views/52/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/14/views/56/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/15/views/60/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/16/views/64/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/17/views/68/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Link Share", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/2/views/8/buckets", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.NoError(t, err)
			db.AssertExists(t, "buckets", map[string]interface{}{
				"project_view_id": 8,
				"created_by_id":   -2,
				"title":           "Lorem Ipsum",
			}, false)
		})
	})
}
