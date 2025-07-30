package webtests

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserExportStatus(t *testing.T) {
	t.Run("no export", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.GetUserExportStatus, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "{}\n", rec.Body.String())
	})

	t.Run("with export", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		_, err := s.Where("id = ?", testuser1.ID).Cols("export_file_id").Update(&user.User{ExportFileID: 1})
		require.NoError(t, err)

		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.GetUserExportStatus, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "\"id\":1")
	})
}
