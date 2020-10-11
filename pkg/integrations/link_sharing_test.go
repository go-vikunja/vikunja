// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLinkSharing(t *testing.T) {

	linkshareRead := &models.LinkSharing{
		ID:          1,
		Hash:        "test1",
		ListID:      1,
		Right:       models.RightRead,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	}

	linkShareWrite := &models.LinkSharing{
		ID:          2,
		Hash:        "test2",
		ListID:      2,
		Right:       models.RightWrite,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	}

	linkShareAdmin := &models.LinkSharing{
		ID:          3,
		Hash:        "test3",
		ListID:      3,
		Right:       models.RightAdmin,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	}

	t.Run("New Link Share", func(t *testing.T) {
		testHandler := webHandlerTest{
			user: &testuser1,
			strFunc: func() handler.CObject {
				return &models.LinkSharing{}
			},
			t: t,
		}
		t.Run("Forbidden", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "20"}, `{"right":0}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "20"}, `{"right":1}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "20"}, `{"right":2}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Read only access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "9"}, `{"right":0}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "9"}, `{"right":1}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "9"}, `{"right":2}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Write access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "10"}, `{"right":0}`)
				assert.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "10"}, `{"right":1}`)
				assert.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "10"}, `{"right":2}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Admin access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "11"}, `{"right":0}`)
				assert.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "11"}, `{"right":1}`)
				assert.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "11"}, `{"right":2}`)
				assert.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
		})
	})

	t.Run("Lists", func(t *testing.T) {
		testHandlerListReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.List{}
			},
			t: t,
		}
		testHandlerListWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.List{}
			},
			t: t,
		}
		testHandlerListAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.List{}
			},
			t: t,
		}

		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Normal", func(t *testing.T) {
				rec, err := testHandlerListReadOnly.testReadAllWithLinkShare(nil, nil)
				assert.NoError(t, err)
				// Should only return the shared list, nothing else
				assert.Contains(t, rec.Body.String(), `Test1`)
				assert.NotContains(t, rec.Body.String(), `Test2`)
				assert.NotContains(t, rec.Body.String(), `Test3`)
				assert.NotContains(t, rec.Body.String(), `Test4`)
				assert.NotContains(t, rec.Body.String(), `Test5`)
			})
			t.Run("Search", func(t *testing.T) {
				rec, err := testHandlerListReadOnly.testReadAllWithLinkShare(url.Values{"s": []string{"est1"}}, nil)
				assert.NoError(t, err)
				// Should only return the shared list, nothing else
				assert.Contains(t, rec.Body.String(), `Test1`)
				assert.NotContains(t, rec.Body.String(), `Test2`)
				assert.NotContains(t, rec.Body.String(), `Test3`)
				assert.NotContains(t, rec.Body.String(), `Test4`)
				assert.NotContains(t, rec.Body.String(), `Test5`)
			})
		})
		t.Run("ReadOne", func(t *testing.T) {
			t.Run("Normal", func(t *testing.T) {
				rec, err := testHandlerListReadOnly.testReadOneWithLinkShare(nil, map[string]string{"list": "1"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
				assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			})
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerListReadOnly.testReadOneWithLinkShare(nil, map[string]string{"list": "9999999"})
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
			})
			t.Run("Rights check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					// List 2, not shared with this token
					_, err := testHandlerListReadOnly.testReadOneWithLinkShare(nil, map[string]string{"list": "2"})
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the right to see this`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					rec, err := testHandlerListReadOnly.testReadOneWithLinkShare(nil, map[string]string{"list": "1"})
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
				})
				t.Run("Shared write", func(t *testing.T) {
					rec, err := testHandlerListWrite.testReadOneWithLinkShare(nil, map[string]string{"list": "2"})
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test2"`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerListAdmin.testReadOneWithLinkShare(nil, map[string]string{"list": "3"})
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test3"`)
				})
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerListReadOnly.testUpdateWithLinkShare(nil, map[string]string{"list": "9999999"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
			})
			t.Run("Rights check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					_, err := testHandlerListReadOnly.testUpdateWithLinkShare(nil, map[string]string{"list": "2"}, `{"title":"TestLoremIpsum"}`)
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerListReadOnly.testUpdateWithLinkShare(nil, map[string]string{"list": "1"}, `{"title":"TestLoremIpsum"}`)
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					rec, err := testHandlerListWrite.testUpdateWithLinkShare(nil, map[string]string{"list": "2"}, `{"title":"TestLoremIpsum"}`)
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerListAdmin.testUpdateWithLinkShare(nil, map[string]string{"list": "3"}, `{"title":"TestLoremIpsum"}`)
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
				})
			})
		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerListReadOnly.testDeleteWithLinkShare(nil, map[string]string{"list": "9999999"})
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
			})
			t.Run("Rights check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					_, err := testHandlerListReadOnly.testDeleteWithLinkShare(nil, map[string]string{"list": "1"})
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerListReadOnly.testDeleteWithLinkShare(nil, map[string]string{"list": "1"})
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					_, err := testHandlerListWrite.testDeleteWithLinkShare(nil, map[string]string{"list": "2"})
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerListAdmin.testDeleteWithLinkShare(nil, map[string]string{"list": "3"})
					assert.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
				})
			})
		})

		// Creating a list should always be forbidden, since users need access to a namespace to create a list
		t.Run("Create", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerListReadOnly.testCreateWithLinkShare(nil, map[string]string{"namespace": "999999"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Rights check", func(t *testing.T) {
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerListReadOnly.testCreateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					_, err := testHandlerListWrite.testCreateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					_, err := testHandlerListAdmin.testCreateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					assert.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
			})
		})

		t.Run("Right Management", func(t *testing.T) {
			t.Run("Users", func(t *testing.T) {
				testHandlerListUserReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.ListUser{}
					},
					t: t,
				}
				testHandlerListUserWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.ListUser{}
					},
					t: t,
				}
				testHandlerListUserAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.ListUser{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						rec, err := testHandlerListUserReadOnly.testReadAllWithLinkShare(nil, map[string]string{"list": "1"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared write", func(t *testing.T) {
						rec, err := testHandlerListUserWrite.testReadAllWithLinkShare(nil, map[string]string{"list": "2"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						rec, err := testHandlerListUserAdmin.testReadAllWithLinkShare(nil, map[string]string{"list": "3"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `"username":"user1"`)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListUserReadOnly.testCreateWithLinkShare(nil, map[string]string{"list": "1"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListUserWrite.testCreateWithLinkShare(nil, map[string]string{"list": "2"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListUserAdmin.testCreateWithLinkShare(nil, map[string]string{"list": "3"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListUserReadOnly.testUpdateWithLinkShare(nil, map[string]string{"list": "1"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListUserWrite.testUpdateWithLinkShare(nil, map[string]string{"list": "2"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListUserAdmin.testUpdateWithLinkShare(nil, map[string]string{"list": "3"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListUserReadOnly.testDeleteWithLinkShare(nil, map[string]string{"list": "1"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListUserWrite.testDeleteWithLinkShare(nil, map[string]string{"list": "2"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListUserAdmin.testDeleteWithLinkShare(nil, map[string]string{"list": "3"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
			})
			t.Run("Teams", func(t *testing.T) {
				testHandlerListTeamReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.TeamList{}
					},
					t: t,
				}
				testHandlerListTeamWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.TeamList{}
					},
					t: t,
				}
				testHandlerListTeamAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.TeamList{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						rec, err := testHandlerListTeamReadOnly.testReadAllWithLinkShare(nil, map[string]string{"list": "1"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared write", func(t *testing.T) {
						rec, err := testHandlerListTeamWrite.testReadAllWithLinkShare(nil, map[string]string{"list": "2"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						rec, err := testHandlerListTeamAdmin.testReadAllWithLinkShare(nil, map[string]string{"list": "3"})
						assert.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListTeamReadOnly.testCreateWithLinkShare(nil, map[string]string{"list": "1"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListTeamWrite.testCreateWithLinkShare(nil, map[string]string{"list": "2"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListTeamAdmin.testCreateWithLinkShare(nil, map[string]string{"list": "3"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListTeamReadOnly.testUpdateWithLinkShare(nil, map[string]string{"list": "1"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListTeamWrite.testUpdateWithLinkShare(nil, map[string]string{"list": "2"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListTeamAdmin.testUpdateWithLinkShare(nil, map[string]string{"list": "3"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerListTeamReadOnly.testDeleteWithLinkShare(nil, map[string]string{"list": "1"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerListTeamWrite.testDeleteWithLinkShare(nil, map[string]string{"list": "2"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerListTeamAdmin.testDeleteWithLinkShare(nil, map[string]string{"list": "3"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
			})
		})
	})

	t.Run("Tasks", func(t *testing.T) {
		testHandlerTaskReadOnlyCollection := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.TaskCollection{}
			},
			t: t,
		}
		testHandlerTaskWriteCollection := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.TaskCollection{}
			},
			t: t,
		}
		testHandlerTaskAdminCollection := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.TaskCollection{}
			},
			t: t,
		}
		testHandlerTaskReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.Task{}
			},
			t: t,
		}
		testHandlerTaskWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.Task{}
			},
			t: t,
		}
		testHandlerTaskAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.Task{}
			},
			t: t,
		}
		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				rec, err := testHandlerTaskReadOnlyCollection.testReadAllWithLinkShare(nil, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `task #1`)
				assert.Contains(t, rec.Body.String(), `task #2`)
				assert.Contains(t, rec.Body.String(), `task #3`)
				assert.Contains(t, rec.Body.String(), `task #4`)
				assert.Contains(t, rec.Body.String(), `task #5`)
				assert.Contains(t, rec.Body.String(), `task #6`)
				assert.Contains(t, rec.Body.String(), `task #7`)
				assert.Contains(t, rec.Body.String(), `task #8`)
				assert.Contains(t, rec.Body.String(), `task #9`)
				assert.Contains(t, rec.Body.String(), `task #10`)
				assert.Contains(t, rec.Body.String(), `task #11`)
				assert.Contains(t, rec.Body.String(), `task #12`)
				assert.NotContains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWriteCollection.testReadAllWithLinkShare(nil, nil)
				assert.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #2`)
				assert.NotContains(t, rec.Body.String(), `task #3`)
				assert.NotContains(t, rec.Body.String(), `task #4`)
				assert.NotContains(t, rec.Body.String(), `task #5`)
				assert.NotContains(t, rec.Body.String(), `task #6`)
				assert.NotContains(t, rec.Body.String(), `task #7`)
				assert.NotContains(t, rec.Body.String(), `task #8`)
				assert.NotContains(t, rec.Body.String(), `task #9`)
				assert.NotContains(t, rec.Body.String(), `task #10`)
				assert.NotContains(t, rec.Body.String(), `task #11`)
				assert.NotContains(t, rec.Body.String(), `task #12`)
				assert.Contains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdminCollection.testReadAllWithLinkShare(nil, nil)
				assert.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #2`)
				assert.NotContains(t, rec.Body.String(), `task #4`)
				assert.NotContains(t, rec.Body.String(), `task #5`)
				assert.NotContains(t, rec.Body.String(), `task #6`)
				assert.NotContains(t, rec.Body.String(), `task #7`)
				assert.NotContains(t, rec.Body.String(), `task #8`)
				assert.NotContains(t, rec.Body.String(), `task #9`)
				assert.NotContains(t, rec.Body.String(), `task #10`)
				assert.NotContains(t, rec.Body.String(), `task #11`)
				assert.NotContains(t, rec.Body.String(), `task #12`)
				assert.NotContains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
				assert.Contains(t, rec.Body.String(), `task #32`)
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTaskReadOnly.testCreateWithLinkShare(nil, map[string]string{"list": "1"}, `{"title":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testCreateWithLinkShare(nil, map[string]string{"list": "2"}, `{"title":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testCreateWithLinkShare(nil, map[string]string{"list": "3"}, `{"title":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTaskReadOnly.testUpdateWithLinkShare(nil, map[string]string{"listtask": "1"}, `{"title":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testUpdateWithLinkShare(nil, map[string]string{"listtask": "13"}, `{"title":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testUpdateWithLinkShare(nil, map[string]string{"listtask": "32"}, `{"title":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTaskReadOnly.testDeleteWithLinkShare(nil, map[string]string{"listtask": "1"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testDeleteWithLinkShare(nil, map[string]string{"listtask": "13"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testDeleteWithLinkShare(nil, map[string]string{"listtask": "32"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})

	t.Run("Teams", func(t *testing.T) {
		testHandlerTeamReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.Team{}
			},
			t: t,
		}
		testHandlerTeamWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.Team{}
			},
			t: t,
		}
		testHandlerTeamAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.Team{}
			},
			t: t,
		}
		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTeamReadOnly.testReadAllWithLinkShare(nil, nil)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testReadAllWithLinkShare(nil, nil)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testReadAllWithLinkShare(nil, nil)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTeamReadOnly.testUpdateWithLinkShare(nil, map[string]string{"team": "1"}, `{"name":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testUpdateWithLinkShare(nil, map[string]string{"team": "2"}, `{"name":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testUpdateWithLinkShare(nil, map[string]string{"team": "3"}, `{"name":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTeamReadOnly.testDeleteWithLinkShare(nil, map[string]string{"team": "1"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testDeleteWithLinkShare(nil, map[string]string{"team": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testDeleteWithLinkShare(nil, map[string]string{"team": "3"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
	})

	t.Run("Linkshare Management", func(t *testing.T) {
		testHandlerLinkShareReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.LinkSharing{}
			},
			t: t,
		}
		testHandlerLinkShareWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.LinkSharing{}
			},
			t: t,
		}
		testHandlerLinkShareAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.LinkSharing{}
			},
			t: t,
		}
		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				rec, err := testHandlerLinkShareReadOnly.testReadAllWithLinkShare(nil, map[string]string{"list": "1"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerLinkShareWrite.testReadAllWithLinkShare(nil, map[string]string{"list": "2"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test2"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerLinkShareAdmin.testReadAllWithLinkShare(nil, map[string]string{"list": "3"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test3"`)
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testCreateWithLinkShare(nil, map[string]string{"list": "1"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testCreateWithLinkShare(nil, map[string]string{"list": "2"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testCreateWithLinkShare(nil, map[string]string{"list": "3"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testUpdateWithLinkShare(nil, map[string]string{"share": "1"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testUpdateWithLinkShare(nil, map[string]string{"share": "2"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testUpdateWithLinkShare(nil, map[string]string{"share": "3"}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testDeleteWithLinkShare(nil, map[string]string{"share": "1"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testDeleteWithLinkShare(nil, map[string]string{"share": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testDeleteWithLinkShare(nil, map[string]string{"share": "3"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
	})

	t.Run("Namespace", func(t *testing.T) {
		testHandlerNamespaceReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.Namespace{}
			},
			t: t,
		}
		testHandlerNamespaceWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.Namespace{}
			},
			t: t,
		}
		testHandlerNamespaceAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.Namespace{}
			},
			t: t,
		}
		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerNamespaceReadOnly.testReadAllWithLinkShare(nil, map[string]string{"namespace": "1"})
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerNamespaceWrite.testReadAllWithLinkShare(nil, map[string]string{"namespace": "2"})
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerNamespaceAdmin.testReadAllWithLinkShare(nil, map[string]string{"namespace": "3"})
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerNamespaceReadOnly.testCreateWithLinkShare(nil, nil, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerNamespaceWrite.testCreateWithLinkShare(nil, nil, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerNamespaceAdmin.testCreateWithLinkShare(nil, nil, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerNamespaceReadOnly.testUpdateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerNamespaceWrite.testUpdateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerNamespaceAdmin.testUpdateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"title":"LoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerNamespaceReadOnly.testDeleteWithLinkShare(nil, map[string]string{"namespace": "1"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerNamespaceWrite.testDeleteWithLinkShare(nil, map[string]string{"namespace": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerNamespaceAdmin.testDeleteWithLinkShare(nil, map[string]string{"namespace": "3"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})

		t.Run("Right Management", func(t *testing.T) {
			t.Run("Users", func(t *testing.T) {
				testHandlerNamespaceUserReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.NamespaceUser{}
					},
					t: t,
				}
				testHandlerNamespaceUserWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.NamespaceUser{}
					},
					t: t,
				}
				testHandlerNamespaceUserAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.NamespaceUser{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceUserReadOnly.testReadAllWithLinkShare(nil, map[string]string{"namespace": "1"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceUserWrite.testReadAllWithLinkShare(nil, map[string]string{"namespace": "2"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceUserAdmin.testReadAllWithLinkShare(nil, map[string]string{"namespace": "3"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceUserReadOnly.testCreateWithLinkShare(nil, nil, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceUserWrite.testCreateWithLinkShare(nil, nil, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceUserAdmin.testCreateWithLinkShare(nil, nil, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceUserReadOnly.testUpdateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceUserWrite.testUpdateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceUserAdmin.testUpdateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"user_id":"user1"}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceUserReadOnly.testDeleteWithLinkShare(nil, map[string]string{"namespace": "1"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceUserWrite.testDeleteWithLinkShare(nil, map[string]string{"namespace": "2"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceUserAdmin.testDeleteWithLinkShare(nil, map[string]string{"namespace": "3"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
			})
			t.Run("Teams", func(t *testing.T) {
				testHandlerNamespaceTeamReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.TeamNamespace{}
					},
					t: t,
				}
				testHandlerNamespaceTeamWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.TeamNamespace{}
					},
					t: t,
				}
				testHandlerNamespaceTeamAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.TeamNamespace{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamReadOnly.testReadAllWithLinkShare(nil, map[string]string{"namespace": "1"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamWrite.testReadAllWithLinkShare(nil, map[string]string{"namespace": "2"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamAdmin.testReadAllWithLinkShare(nil, map[string]string{"namespace": "3"})
						assert.Error(t, err)
						assertHandlerErrorCode(t, err, models.ErrCodeNeedToHaveNamespaceReadAccess)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamReadOnly.testCreateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamWrite.testCreateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamAdmin.testCreateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamReadOnly.testUpdateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamWrite.testUpdateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamAdmin.testUpdateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"team_id":1}`)
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamReadOnly.testDeleteWithLinkShare(nil, map[string]string{"namespace": "1"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamWrite.testDeleteWithLinkShare(nil, map[string]string{"namespace": "2"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerNamespaceTeamAdmin.testDeleteWithLinkShare(nil, map[string]string{"namespace": "3"})
						assert.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
			})
		})
	})
}
