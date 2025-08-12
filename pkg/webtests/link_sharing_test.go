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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkSharing(t *testing.T) {

	linkshareRead := &models.LinkSharing{
		ID:          1,
		Hash:        "test1",
		ProjectID:   1,
		Permission:  models.PermissionRead,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	}

	linkShareWrite := &models.LinkSharing{
		ID:          2,
		Hash:        "test2",
		ProjectID:   2,
		Permission:  models.PermissionWrite,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	}

	linkShareAdmin := &models.LinkSharing{
		ID:          3,
		Hash:        "test3",
		ProjectID:   3,
		Permission:  models.PermissionAdmin,
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
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "20"}, `{"permission":0}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "20"}, `{"permission":1}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "20"}, `{"permission":2}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Read only access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "9"}, `{"permission":0}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "9"}, `{"permission":1}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "9"}, `{"permission":2}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Write access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "10"}, `{"permission":0}`)
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "10"}, `{"permission":1}`)
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "10"}, `{"permission":2}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Admin access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "11"}, `{"permission":0}`)
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "11"}, `{"permission":1}`)
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				req, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "11"}, `{"permission":2}`)
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
		})
	})

	t.Run("Projects", func(t *testing.T) {
		testHandlerProjectReadOnly := webHandlerTest{
			linkShare: linkshareRead,
			strFunc: func() handler.CObject {
				return &models.Project{}
			},
			t: t,
		}
		testHandlerProjectWrite := webHandlerTest{
			linkShare: linkShareWrite,
			strFunc: func() handler.CObject {
				return &models.Project{}
			},
			t: t,
		}
		testHandlerProjectAdmin := webHandlerTest{
			linkShare: linkShareAdmin,
			strFunc: func() handler.CObject {
				return &models.Project{}
			},
			t: t,
		}

		t.Run("ReadAll", func(t *testing.T) {
			t.Run("Normal", func(t *testing.T) {
				rec, err := testHandlerProjectReadOnly.testReadAllWithLinkShare(nil, nil)
				require.NoError(t, err)
				// Should only return the shared project, nothing else
				assert.Contains(t, rec.Body.String(), `Test1`)
				assert.NotContains(t, rec.Body.String(), `Test2`)
				assert.NotContains(t, rec.Body.String(), `Test3`)
				assert.NotContains(t, rec.Body.String(), `Test4`)
				assert.NotContains(t, rec.Body.String(), `Test5`)
			})
			t.Run("Search", func(t *testing.T) {
				rec, err := testHandlerProjectReadOnly.testReadAllWithLinkShare(url.Values{"s": []string{"est1"}}, nil)
				require.NoError(t, err)
				// Should only return the shared project, nothing else
				assert.Contains(t, rec.Body.String(), `Test1`)
				assert.NotContains(t, rec.Body.String(), `Test2`)
				assert.NotContains(t, rec.Body.String(), `Test3`)
				assert.NotContains(t, rec.Body.String(), `Test4`)
				assert.NotContains(t, rec.Body.String(), `Test5`)
			})
		})
		t.Run("ReadOne", func(t *testing.T) {
			t.Run("Normal", func(t *testing.T) {
				rec, err := testHandlerProjectReadOnly.testReadOneWithLinkShare(nil, map[string]string{"project": "1"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
				assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			})
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerProjectReadOnly.testReadOneWithLinkShare(nil, map[string]string{"project": "9999999"})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
			})
			t.Run("Permissions check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					// Project 2, not shared with this token
					_, err := testHandlerProjectReadOnly.testReadOneWithLinkShare(nil, map[string]string{"project": "2"})
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the permission to see this`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					rec, err := testHandlerProjectReadOnly.testReadOneWithLinkShare(nil, map[string]string{"project": "1"})
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
				})
				t.Run("Shared write", func(t *testing.T) {
					rec, err := testHandlerProjectWrite.testReadOneWithLinkShare(nil, map[string]string{"project": "2"})
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test2"`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerProjectAdmin.testReadOneWithLinkShare(nil, map[string]string{"project": "3"})
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"Test3"`)
				})
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerProjectReadOnly.testUpdateWithLinkShare(nil, map[string]string{"project": "9999999"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
			})
			t.Run("Permissions check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					_, err := testHandlerProjectReadOnly.testUpdateWithLinkShare(nil, map[string]string{"project": "2"}, `{"title":"TestLoremIpsum"}`)
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerProjectReadOnly.testUpdateWithLinkShare(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum"}`)
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					rec, err := testHandlerProjectWrite.testUpdateWithLinkShare(nil, map[string]string{"project": "2"}, `{"title":"TestLoremIpsum","namespace_id":1}`)
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerProjectAdmin.testUpdateWithLinkShare(nil, map[string]string{"project": "3"}, `{"title":"TestLoremIpsum","namespace_id":2}`)
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
				})
			})
		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerProjectReadOnly.testDeleteWithLinkShare(nil, map[string]string{"project": "9999999"})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
			})
			t.Run("Permissions check", func(t *testing.T) {
				t.Run("Forbidden", func(t *testing.T) {
					_, err := testHandlerProjectReadOnly.testDeleteWithLinkShare(nil, map[string]string{"project": "1"})
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerProjectReadOnly.testDeleteWithLinkShare(nil, map[string]string{"project": "1"})
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					_, err := testHandlerProjectWrite.testDeleteWithLinkShare(nil, map[string]string{"project": "2"})
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					rec, err := testHandlerProjectAdmin.testDeleteWithLinkShare(nil, map[string]string{"project": "3"})
					require.NoError(t, err)
					assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
				})
			})
		})

		// Creating a project should always be forbidden
		t.Run("Create", func(t *testing.T) {
			t.Run("Nonexisting", func(t *testing.T) {
				_, err := testHandlerProjectReadOnly.testCreateWithLinkShare(nil, nil, `{"title":"Lorem"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Permissions check", func(t *testing.T) {
				t.Run("Shared readonly", func(t *testing.T) {
					_, err := testHandlerProjectReadOnly.testCreateWithLinkShare(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared write", func(t *testing.T) {
					_, err := testHandlerProjectWrite.testCreateWithLinkShare(nil, map[string]string{"namespace": "2"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
				t.Run("Shared admin", func(t *testing.T) {
					_, err := testHandlerProjectAdmin.testCreateWithLinkShare(nil, map[string]string{"namespace": "3"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
					require.Error(t, err)
					assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
				})
			})
		})

		t.Run("Permission Management", func(t *testing.T) {
			t.Run("Users", func(t *testing.T) {
				testHandlerProjectUserReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.ProjectUser{}
					},
					t: t,
				}
				testHandlerProjectUserWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.ProjectUser{}
					},
					t: t,
				}
				testHandlerProjectUserAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.ProjectUser{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						rec, err := testHandlerProjectUserReadOnly.testReadAllWithLinkShare(nil, map[string]string{"project": "1"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared write", func(t *testing.T) {
						rec, err := testHandlerProjectUserWrite.testReadAllWithLinkShare(nil, map[string]string{"project": "2"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						rec, err := testHandlerProjectUserAdmin.testReadAllWithLinkShare(nil, map[string]string{"project": "3"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `"username":"user1"`)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectUserReadOnly.testCreateWithLinkShare(nil, map[string]string{"project": "1"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectUserWrite.testCreateWithLinkShare(nil, map[string]string{"project": "2"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectUserAdmin.testCreateWithLinkShare(nil, map[string]string{"project": "3"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectUserReadOnly.testUpdateWithLinkShare(nil, map[string]string{"project": "1"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectUserWrite.testUpdateWithLinkShare(nil, map[string]string{"project": "2"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectUserAdmin.testUpdateWithLinkShare(nil, map[string]string{"project": "3"}, `{"user_id":"user1"}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectUserReadOnly.testDeleteWithLinkShare(nil, map[string]string{"project": "1"})
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectUserWrite.testDeleteWithLinkShare(nil, map[string]string{"project": "2"})
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectUserAdmin.testDeleteWithLinkShare(nil, map[string]string{"project": "3"})
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
			})
			t.Run("Teams", func(t *testing.T) {
				testHandlerProjectTeamReadOnly := webHandlerTest{
					linkShare: linkshareRead,
					strFunc: func() handler.CObject {
						return &models.TeamProject{}
					},
					t: t,
				}
				testHandlerProjectTeamWrite := webHandlerTest{
					linkShare: linkShareWrite,
					strFunc: func() handler.CObject {
						return &models.TeamProject{}
					},
					t: t,
				}
				testHandlerProjectTeamAdmin := webHandlerTest{
					linkShare: linkShareAdmin,
					strFunc: func() handler.CObject {
						return &models.TeamProject{}
					},
					t: t,
				}
				t.Run("ReadAll", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						rec, err := testHandlerProjectTeamReadOnly.testReadAllWithLinkShare(nil, map[string]string{"project": "1"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared write", func(t *testing.T) {
						rec, err := testHandlerProjectTeamWrite.testReadAllWithLinkShare(nil, map[string]string{"project": "2"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `[]`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						rec, err := testHandlerProjectTeamAdmin.testReadAllWithLinkShare(nil, map[string]string{"project": "3"})
						require.NoError(t, err)
						assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
					})
				})
				t.Run("Create", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectTeamReadOnly.testCreateWithLinkShare(nil, map[string]string{"project": "1"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectTeamWrite.testCreateWithLinkShare(nil, map[string]string{"project": "2"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectTeamAdmin.testCreateWithLinkShare(nil, map[string]string{"project": "3"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
				})
				t.Run("Update", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectTeamReadOnly.testUpdateWithLinkShare(nil, map[string]string{"project": "1"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectTeamWrite.testUpdateWithLinkShare(nil, map[string]string{"project": "2"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectTeamAdmin.testUpdateWithLinkShare(nil, map[string]string{"project": "3"}, `{"team_id":1}`)
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})

				})
				t.Run("Delete", func(t *testing.T) {
					t.Run("Shared readonly", func(t *testing.T) {
						_, err := testHandlerProjectTeamReadOnly.testDeleteWithLinkShare(nil, map[string]string{"project": "1"})
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared write", func(t *testing.T) {
						_, err := testHandlerProjectTeamWrite.testDeleteWithLinkShare(nil, map[string]string{"project": "2"})
						require.Error(t, err)
						assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
					})
					t.Run("Shared admin", func(t *testing.T) {
						_, err := testHandlerProjectTeamAdmin.testDeleteWithLinkShare(nil, map[string]string{"project": "3"})
						require.Error(t, err)
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
				require.NoError(t, err)
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
				require.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #2`)
				assert.NotContains(t, rec.Body.String(), `task #3"`)
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
				require.NoError(t, err)
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
				_, err := testHandlerTaskReadOnly.testCreateWithLinkShare(nil, map[string]string{"project": "1"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testCreateWithLinkShare(nil, map[string]string{"project": "2"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testCreateWithLinkShare(nil, map[string]string{"project": "3"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTaskReadOnly.testUpdateWithLinkShare(nil, map[string]string{"projecttask": "1"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testUpdateWithLinkShare(nil, map[string]string{"projecttask": "13"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testUpdateWithLinkShare(nil, map[string]string{"projecttask": "32"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTaskReadOnly.testDeleteWithLinkShare(nil, map[string]string{"projecttask": "1"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerTaskWrite.testDeleteWithLinkShare(nil, map[string]string{"projecttask": "13"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerTaskAdmin.testDeleteWithLinkShare(nil, map[string]string{"projecttask": "32"})
				require.NoError(t, err)
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
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testReadAllWithLinkShare(nil, nil)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testReadAllWithLinkShare(nil, nil)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTeamReadOnly.testUpdateWithLinkShare(nil, map[string]string{"team": "1"}, `{"name":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testUpdateWithLinkShare(nil, map[string]string{"team": "2"}, `{"name":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testUpdateWithLinkShare(nil, map[string]string{"team": "3"}, `{"name":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerTeamReadOnly.testDeleteWithLinkShare(nil, map[string]string{"team": "1"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerTeamWrite.testDeleteWithLinkShare(nil, map[string]string{"team": "2"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerTeamAdmin.testDeleteWithLinkShare(nil, map[string]string{"team": "3"})
				require.Error(t, err)
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
				rec, err := testHandlerLinkShareReadOnly.testReadAllWithLinkShare(nil, map[string]string{"project": "1"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			})
			t.Run("Shared write", func(t *testing.T) {
				rec, err := testHandlerLinkShareWrite.testReadAllWithLinkShare(nil, map[string]string{"project": "2"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test2"`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				rec, err := testHandlerLinkShareAdmin.testReadAllWithLinkShare(nil, map[string]string{"project": "3"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hash":"test3"`)
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testCreateWithLinkShare(nil, map[string]string{"project": "1"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testCreateWithLinkShare(nil, map[string]string{"project": "2"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testCreateWithLinkShare(nil, map[string]string{"project": "3"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testUpdateWithLinkShare(nil, map[string]string{"share": "1"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testUpdateWithLinkShare(nil, map[string]string{"share": "2"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testUpdateWithLinkShare(nil, map[string]string{"share": "3"}, `{}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("Shared readonly", func(t *testing.T) {
				_, err := testHandlerLinkShareReadOnly.testDeleteWithLinkShare(nil, map[string]string{"share": "1"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared write", func(t *testing.T) {
				_, err := testHandlerLinkShareWrite.testDeleteWithLinkShare(nil, map[string]string{"share": "2"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared admin", func(t *testing.T) {
				_, err := testHandlerLinkShareAdmin.testDeleteWithLinkShare(nil, map[string]string{"share": "3"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
	})
}
