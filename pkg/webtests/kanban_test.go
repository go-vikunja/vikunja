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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucket(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Bucket{}
		},
		t: t,
	}
	testHandlerLinkShareWrite := webHandlerTest{
		linkShare: &models.LinkSharing{
			ID:          2,
			Hash:        "test2",
			ProjectID:   2,
			Permission:  models.PermissionWrite,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		},
		strFunc: func() handler.CObject {
			return &models.Bucket{}
		},
		t: t,
	}
	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, map[string]string{
				"project": "1",
				"view":    "4",
			})
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
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
				"bucket":  "1",
				"project": "1",
				"view":    "4",
			}, `{"title":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
		})
		t.Run("Nonexisting Bucket", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{
				"bucket":  "9999",
				"project": "1",
				"view":    "4",
			}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeBucketDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{
				"bucket":  "1",
				"project": "1",
				"view":    "4",
			}, `{"title":""}`)
			require.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "5",
					"project": "20",
					"view":    "80",
				}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "6",
					"project": "6",
					"view":    "24",
				}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "7",
					"project": "7",
					"view":    "28",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "8",
					"project": "8",
					"view":    "32",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "9",
					"project": "9",
					"view":    "36",
				}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "10",
					"project": "10",
					"view":    "40",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "11",
					"project": "11",
					"view":    "44",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "12",
					"project": "12",
					"view":    "48",
				}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "13",
					"project": "13",
					"view":    "52",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "14",
					"project": "14",
					"view":    "56",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "15",
					"project": "15",
					"view":    "60",
				}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "16",
					"project": "16",
					"view":    "64",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{
					"bucket":  "17",
					"project": "17",
					"view":    "68",
				}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
				"project": "1",
				"bucket":  "1",
				"view":    "4",
			})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"bucket": "999"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeBucketDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "20", "bucket": "5"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "6", "bucket": "6"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "7",
					"bucket":  "7",
					"view":    "28",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "8",
					"bucket":  "8",
					"view":    "32",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "9",
					"bucket":  "9",
					"view":    "36",
				})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "10",
					"bucket":  "10",
					"view":    "40",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "11",
					"bucket":  "11",
					"view":    "44",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "12",
					"bucket":  "12",
					"view":    "48",
				})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "13",
					"bucket":  "13",
					"view":    "52",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "14",
					"bucket":  "14",
					"view":    "56",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "15",
					"bucket":  "15",
					"view":    "60",
				})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "16",
					"bucket":  "16",
					"view":    "64",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{
					"project": "17",
					"bucket":  "17",
					"view":    "68",
				})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{
				"project": "1",
				"view":    "3",
			}, `{"title":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
		})
		t.Run("Nonexistent project", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{
				"project": "9999",
				"view":    "1",
			}, `{"title":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectViewDoesNotExist)
		})
		t.Run("Nonexistent view", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{
				"project": "1",
				"view":    "9999",
			}, `{"title":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectViewDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "20",
					"view":    "80",
				}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "6",
					"view":    "24",
				}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "7",
					"view":    "28",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "8",
					"view":    "32",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "9",
					"view":    "36",
				}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "10",
					"view":    "40",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "11",
					"view":    "44",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "12",
					"view":    "48",
				}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "13",
					"view":    "52",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "14",
					"view":    "56",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "15",
					"view":    "60",
				}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "16",
					"view":    "64",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{
					"project": "17",
					"view":    "68",
				}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Link Share", func(t *testing.T) {
			rec, err := testHandlerLinkShareWrite.testCreateWithLinkShare(nil, map[string]string{
				"project": "2",
				"view":    "8",
			}, `{"title":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			db.AssertExists(t, "buckets", map[string]interface{}{
				"project_view_id": 8,
				"created_by_id":   -2,
				"title":           "Lorem Ipsum",
			}, false)
		})
	})
}
