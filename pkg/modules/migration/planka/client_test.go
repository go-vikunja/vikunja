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

package planka

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePlanka records requests and answers like a Planka v2 server.
type fakePlanka struct {
	t *testing.T
	// validAPIKey / validJWT are accepted credentials; empty disables the method.
	validAPIKey string
	validJWT    string
	// loginJWT is returned by POST /api/access-tokens; pendingStep makes login return that step instead.
	loginJWT    string
	pendingStep string
	// meStatus, when set, is the status /api/users/me answers with instead of the user.
	meStatus int
	// fixtures maps a path to a testdata file served for the first page of that path.
	fixtures map[string]string

	mu       sync.Mutex
	requests []*http.Request
	deleted  int
}

func (f *fakePlanka) authenticated(r *http.Request) bool {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return f.validJWT != "" && auth == "Bearer "+f.validJWT
	}
	if key := r.Header.Get("X-Api-Key"); key != "" {
		return f.validAPIKey != "" && key == f.validAPIKey
	}
	return false
}

func (f *fakePlanka) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.requests = append(f.requests, r.Clone(r.Context()))
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/access-tokens":
			var body map[string]any
			assert.NoError(f.t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(f.t, "user@example.com", body["emailOrUsername"])
			assert.Equal(f.t, "secret", body["password"])
			_, hasHTTPOnly := body["withHttpOnlyToken"]
			assert.False(f.t, hasHTTPOnly, "withHttpOnlyToken must not be sent")
			if f.pendingStep != "" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"E_FORBIDDEN","pendingToken":"pending","message":"extra step required","step":"` + f.pendingStep + `"}`))
				return
			}
			if f.loginJWT == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":"E_UNAUTHORIZED","message":"Invalid credentials"}`))
				return
			}
			_, _ = w.Write([]byte(`{"item":"` + f.loginJWT + `"}`))
			return
		case r.Method == http.MethodDelete && r.URL.Path == "/api/access-tokens/me":
			if !f.authenticated(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			f.mu.Lock()
			f.deleted++
			f.mu.Unlock()
			_, _ = w.Write([]byte(`{"item":"x"}`))
			return
		}

		if !f.authenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":"E_UNAUTHORIZED","message":"Access token is missing, invalid or expired"}`))
			return
		}

		switch r.URL.Path {
		case "/api/users/me":
			if f.meStatus != 0 {
				w.WriteHeader(f.meStatus)
				return
			}
			_, _ = w.Write([]byte(`{"item":{"id":"1","name":"Me","username":"me"},"included":{}}`))
		case "/api/cards/42/comments":
			// 50 items on the first page, 1 on the second, none on the third
			switch before := r.URL.Query().Get("beforeId"); before {
			case "":
				items := []string{}
				for i := 100; i > 50; i-- {
					items = append(items, `{"id":"`+strconv.Itoa(i)+`","text":"c`+strconv.Itoa(i)+`","cardId":"42","userId":"1"}`)
				}
				_, _ = w.Write([]byte(`{"items":[` + strings.Join(items, ",") + `],"included":{"users":[]}}`))
			case "51":
				_, _ = w.Write([]byte(`{"items":[{"id":"50","text":"c50","cardId":"42","userId":"1"}],"included":{"users":[]}}`))
			default:
				assert.Equal(f.t, "50", before)
				_, _ = w.Write([]byte(`{"items":[],"included":{"users":[]}}`))
			}
		case "/api/cards/43/comments":
			// the second page fails, the caller keeps the first one
			if r.URL.Query().Get("beforeId") != "" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"id":"3","text":"c3","cardId":"43","userId":"1"},{"id":"2","text":"c2","cardId":"43","userId":"1"}],"included":{"users":[]}}`))
		case "/api/lists/77/cards":
			q := r.URL.Query()
			switch q.Get("before[id]") {
			case "":
				items := []string{}
				for i := 100; i > 50; i-- {
					items = append(items, `{"id":"`+strconv.Itoa(i)+`","name":"c`+strconv.Itoa(i)+`","listId":"77","listChangedAt":"2024-01-0`+strconv.Itoa(i%9+1)+`T00:00:00.000Z"}`)
				}
				_, _ = w.Write([]byte(`{"items":[` + strings.Join(items, ",") + `],"included":{"users":[],"cardLabels":[{"cardId":"51","labelId":"1"}]}}`))
			case "51":
				assert.Equal(f.t, "2024-01-07T00:00:00Z", q.Get("before[listChangedAt]"))
				_, _ = w.Write([]byte(`{"items":[{"id":"50","name":"c50","listId":"77","listChangedAt":"2024-01-02T00:00:00.000Z"}],"included":{"users":[],"cardLabels":[{"cardId":"50","labelId":"1"}]}}`))
			default:
				assert.Equal(f.t, "50", q.Get("before[id]"))
				assert.Equal(f.t, "2024-01-02T00:00:00Z", q.Get("before[listChangedAt]"), "planka rejects a cursor without the timestamp")
				_, _ = w.Write([]byte(`{"items":[],"included":{}}`))
			}
		case "/api/lists/78/cards":
			assert.False(f.t, r.URL.Query().Has("before[id]"), "a card without listChangedAt cannot be paged past")
			_, _ = w.Write([]byte(`{"items":[{"id":"9","name":"c9","listId":"78","listChangedAt":null}],"included":{}}`))
		case "/attachments/7/download/big.bin":
			_, _ = w.Write(make([]byte, 3*1024*1024))
		case "/attachments/8/download/small.bin":
			_, _ = w.Write([]byte("hello"))
		case "/attachments/3000/download/cover.png":
			_, _ = w.Write([]byte("PNG!"))
		default:
			f.serveFixture(w, r)
		}
	})
}

// serveFixture answers with a testdata file. Every fixture is a single page, so paged requests get nothing.
func (f *fakePlanka) serveFixture(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Has("before[id]") || q.Has("beforeId") {
		_, _ = w.Write([]byte(`{"items":[],"included":{}}`))
		return
	}

	name, has := f.fixtures[r.URL.Path]
	if !has {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if !assert.NoError(f.t, err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(content)
}

func newFake(t *testing.T) (*fakePlanka, *httptest.Server) {
	f := &fakePlanka{t: t}
	srv := httptest.NewServer(f.handler())
	t.Cleanup(srv.Close)
	return f, srv
}

func (f *fakePlanka) requestsTo(path string) []*http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*http.Request{}
	for _, r := range f.requests {
		if r.URL.Path == path {
			out = append(out, r)
		}
	}
	return out
}

func TestNewClientNormalisesURL(t *testing.T) {
	for _, in := range []string{
		"https://planka.example.com",
		"https://planka.example.com/",
		"https://planka.example.com/api",
		"https://planka.example.com/api/",
		"  https://planka.example.com//  ",
	} {
		c, err := newClient(in)
		require.NoError(t, err, in)
		assert.Equal(t, "https://planka.example.com", c.baseURL, in)
	}

	c, err := newClient("https://example.com/planka/api/")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/planka", c.baseURL)

	c, err = newClient("https://alice:hunter2@planka.example.com/api")
	require.NoError(t, err)
	assert.Equal(t, "https://planka.example.com", c.baseURL, "userinfo is stripped")

	_, err = newClient("")
	var errInvalid *ErrInvalidConfig
	require.ErrorAs(t, err, &errInvalid)

	for _, in := range []string{"planka.example.com", "127.0.0.1:3999", "ftp://planka.example.com", "https://"} {
		_, err := newClient(in)
		var errNoPlanka *ErrNoPlankaAtURL
		assert.ErrorAs(t, err, &errNoPlanka, in)
	}
}

func TestClientDoesNotFollowCrossHostRedirects(t *testing.T) {
	other, otherSrv := newFake(t)
	other.validAPIKey = "key"
	other.loginJWT = "jwt"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", otherSrv.URL+r.URL.Path)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	t.Cleanup(srv.Close)

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	var errNoPlanka *ErrNoPlankaAtURL
	require.ErrorAs(t, c.login(t.Context(), "key", "", ""), &errNoPlanka)
	require.ErrorAs(t, c.login(t.Context(), "", "user@example.com", "secret"), &errNoPlanka)
	assert.Empty(t, other.requestsTo("/api/users/me"), "the api key must not be replayed on another host")
	assert.Empty(t, other.requestsTo("/api/access-tokens"), "the password must not be replayed on another host")

	c.apiKey = "key"
	err = c.get("/api/projects", nil, &projectsResponse{})
	require.ErrorIs(t, err, migration.ErrRedirectRefused)
	assert.Empty(t, other.requestsTo("/api/projects"), "the api key must not be replayed on another host")
}

func TestDownloadFollowsCrossHostRedirectsWithoutCredentials(t *testing.T) {
	// planka answers with a 302 to the bucket when attachments live on object storage
	var (
		mu   sync.Mutex
		last *http.Request
	)
	storageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		last = r.Clone(r.Context())
		mu.Unlock()
		_, _ = w.Write([]byte("hello"))
	}))
	t.Cleanup(storageSrv.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", storageSrv.URL+r.URL.Path)
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	for _, tc := range []struct{ name, apiKey, jwt string }{
		{name: "api key", apiKey: "key"},
		{name: "jwt", jwt: "jwt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := newClient(srv.URL)
			require.NoError(t, err)
			c.apiKey, c.jwt = tc.apiKey, tc.jwt

			buf, err := c.download("8", "small.bin")
			require.NoError(t, err)
			assert.Equal(t, "hello", buf.String())

			mu.Lock()
			defer mu.Unlock()
			require.NotNil(t, last)
			assert.Equal(t, "/attachments/8/download/small.bin", last.URL.Path)
			assert.Empty(t, last.Header.Get("X-Api-Key"), "credentials must not travel to another host")
			assert.Empty(t, last.Header.Get("Cookie"))
			assert.Empty(t, last.Header.Get("Authorization"))
		})
	}
}

func TestRedirectPolicy(t *testing.T) {
	newReq := func(target string) *http.Request {
		req, err := http.NewRequest(http.MethodGet, target, nil) //nolint:noctx // no request is sent
		require.NoError(t, err)
		return req
	}

	origin := func(target string) *url.URL {
		u, err := url.Parse(target)
		require.NoError(t, err)
		return u
	}

	policy := redirectPolicy(origin("https://planka.example.com"))
	require.NoError(t, policy(newReq("https://PLANKA.example.com/api/users/me"), nil), "the host compare is case insensitive")
	require.NoError(t, policy(newReq("https://planka.example.com:443/api/users/me"), nil), "the default port is not a different host")
	require.ErrorIs(t, policy(newReq("http://planka.example.com/api/users/me"), nil), migration.ErrRedirectRefused, "https must not be downgraded")
	require.ErrorIs(t, policy(newReq("https://evil.example.com/api/users/me"), nil), migration.ErrRedirectRefused)
	require.ErrorIs(t, policy(newReq("https://planka.example.com/"), make([]*http.Request, 10)), migration.ErrRedirectRefused)

	policy = redirectPolicy(origin("http://planka.example.com:80"))
	require.NoError(t, policy(newReq("http://planka.example.com/api/users/me"), nil))

	downloads := downloadRedirectPolicy(origin("https://planka.example.com"))
	req := newReq("https://bucket.example.com/attachments/1")
	req.Header.Set("X-Api-Key", "key")
	req.Header.Set("Cookie", "accessToken=jwt")
	req.Header.Set("Authorization", "Bearer jwt")
	require.NoError(t, downloads(req, nil), "another host is allowed for downloads")
	assert.Empty(t, req.Header.Get("X-Api-Key"))
	assert.Empty(t, req.Header.Get("Cookie"))
	assert.Empty(t, req.Header.Get("Authorization"))

	require.ErrorIs(t, downloads(newReq("http://bucket.example.com/attachments/1"), nil), migration.ErrRedirectRefused, "https must not be downgraded")
	require.ErrorIs(t, downloads(newReq("https://bucket.example.com/"), make([]*http.Request, 10)), migration.ErrRedirectRefused)
}

func TestLoginRejectsEmptyUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/access-tokens" {
			_, _ = w.Write([]byte(`{"item":"jwt"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	var errNoPlanka *ErrNoPlankaAtURL
	require.ErrorAs(t, c.login(t.Context(), "key", "", ""), &errNoPlanka)
	require.ErrorAs(t, c.login(t.Context(), "", "user@example.com", "secret"), &errNoPlanka)
}

func TestCheckCredentialsIsSSRFSafe(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	config.OutgoingRequestsAllowNonRoutableIPs.Set(false)
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(true) })

	m := &Migrator{URL: srv.URL, Token: "key"}
	var errNoPlanka *ErrNoPlankaAtURL
	require.ErrorAs(t, m.CheckCredentials(), &errNoPlanka)
}

func TestLoginWithAPIKey(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key123"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key123", "", ""))
	assert.Equal(t, "1", c.currentUserID)

	me := &plankaUserResponse{}
	require.NoError(t, c.get("/api/users/me", nil, me))

	for _, r := range f.requestsTo("/api/users/me") {
		assert.Equal(t, "key123", r.Header.Get("X-Api-Key"))
		assert.Empty(t, r.Header.Get("Authorization"))
	}
	assert.Equal(t, "key123", c.downloadHeaders().Get("X-Api-Key"))
	assert.Empty(t, c.downloadHeaders().Get("Cookie"))
}

func TestLoginWithBearerToken(t *testing.T) {
	f, srv := newFake(t)
	f.validJWT = "jwt123"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "jwt123", "", ""))

	me := &plankaUserResponse{}
	require.NoError(t, c.get("/api/users/me", nil, me))

	reqs := f.requestsTo("/api/users/me")
	require.Len(t, reqs, 3)
	// first probe uses the api key header only, the rest bearer only
	assert.Equal(t, "jwt123", reqs[0].Header.Get("X-Api-Key"))
	assert.Empty(t, reqs[0].Header.Get("Authorization"))
	for _, r := range reqs[1:] {
		assert.Equal(t, "Bearer jwt123", r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Api-Key"))
	}
	assert.Equal(t, "accessToken=jwt123", c.downloadHeaders().Get("Cookie"))

	// a token we did not create is not deleted on logout
	c.logout(t.Context())
	assert.Equal(t, 0, f.deleted)
}

func TestLoginTokenRejected(t *testing.T) {
	_, srv := newFake(t)

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	err = c.login(t.Context(), "nope", "", "")
	var errInvalid *ErrInvalidCredentials
	require.ErrorAs(t, err, &errInvalid)
}

func TestLoginNoPlankaAtURL(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	var errNoPlanka *ErrNoPlankaAtURL
	require.ErrorAs(t, c.login(t.Context(), "key", "", ""), &errNoPlanka)
	require.ErrorAs(t, c.login(t.Context(), "", "user@example.com", "secret"), &errNoPlanka)

	// nothing listening
	c, err = newClient("http://127.0.0.1:1")
	require.NoError(t, err)
	require.ErrorAs(t, c.login(t.Context(), "key", "", ""), &errNoPlanka)
}

func TestLoginWithPassword(t *testing.T) {
	f, srv := newFake(t)
	f.loginJWT = "sessionjwt"
	f.validJWT = "sessionjwt"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "", "user@example.com", "secret"))
	assert.Equal(t, "1", c.currentUserID)

	for _, r := range f.requestsTo("/api/users/me") {
		assert.Equal(t, "Bearer sessionjwt", r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Api-Key"))
	}

	c.logout(t.Context())
	assert.Equal(t, 1, f.deleted)
}

func TestLoginDeletesSessionWhenTheLoginFailsLater(t *testing.T) {
	f, srv := newFake(t)
	f.loginJWT = "sessionjwt"
	f.validJWT = "sessionjwt"
	f.meStatus = http.StatusInternalServerError

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	var errNoPlanka *ErrNoPlankaAtURL
	require.ErrorAs(t, c.login(t.Context(), "", "user@example.com", "secret"), &errNoPlanka)
	assert.Equal(t, 1, f.deleted, "the session we created must not leak")
}

func TestLoginWithPasswordRejected(t *testing.T) {
	_, srv := newFake(t)

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	err = c.login(t.Context(), "", "user@example.com", "secret")
	var errInvalid *ErrInvalidCredentials
	require.ErrorAs(t, err, &errInvalid)
}

func TestLoginWithPasswordTOTP(t *testing.T) {
	f, srv := newFake(t)
	f.pendingStep = "verify-totp"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	err = c.login(t.Context(), "", "user@example.com", "secret")
	var errStep *ErrLoginStepRequired
	require.ErrorAs(t, err, &errStep)
	assert.Equal(t, "verify-totp", errStep.Step)
}

func TestLoginWithPasswordAcceptTerms(t *testing.T) {
	f, srv := newFake(t)
	f.pendingStep = "accept-terms"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	err = c.login(t.Context(), "", "user@example.com", "secret")
	var errStep *ErrLoginStepRequired
	require.ErrorAs(t, err, &errStep)
	assert.Equal(t, "accept-terms", errStep.Step)
}

func TestLoginMissingCredentials(t *testing.T) {
	_, srv := newFake(t)
	c, err := newClient(srv.URL)
	require.NoError(t, err)

	var errInvalid *ErrInvalidConfig
	require.ErrorAs(t, c.login(t.Context(), "", "", ""), &errInvalid)
	require.ErrorAs(t, c.login(t.Context(), "", "user", ""), &errInvalid)
}

func TestFetchCommentsPaginates(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	comments, _, err := fetchComments(c, "42")
	require.NoError(t, err)
	require.Len(t, comments, 51)
	assert.Equal(t, "50", comments[0].ID, "oldest first")
	assert.Equal(t, "100", comments[50].ID)
	assert.Len(t, f.requestsTo("/api/cards/42/comments"), 3)
}

func TestFetchCommentsPartialResultIsOldestFirst(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	comments, _, err := fetchComments(c, "43")
	require.Error(t, err)
	require.Len(t, comments, 2)
	assert.Equal(t, "2", comments[0].ID, "oldest first, also for a partial result")
	assert.Equal(t, "3", comments[1].ID)
}

func TestFetchArchivedCardsPaginates(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	bd := &plankaBoardData{}
	require.NoError(t, fetchArchivedCards(c, "77", bd))
	assert.Len(t, bd.Cards, 51)
	assert.Len(t, bd.CardLabels, 2, "included data of every page is merged")
	assert.Len(t, f.requestsTo("/api/lists/77/cards"), 3)
}

func TestFetchArchivedCardsStopsWithoutListChangedAt(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	bd := &plankaBoardData{}
	require.NoError(t, fetchArchivedCards(c, "78", bd))
	assert.Len(t, bd.Cards, 1)
	assert.Len(t, f.requestsTo("/api/lists/78/cards"), 1, "paging stops instead of sending a cursor planka rejects")
}

func TestDownloadCapsSize(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"

	oldMax := config.FilesMaxSize.GetString()
	config.FilesMaxSize.Set("1MB")
	require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
	t.Cleanup(func() {
		config.FilesMaxSize.Set(oldMax)
		_ = config.SetMaxFileSizeMBytesFromString(oldMax)
	})

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	_, err = c.download("7", "big.bin")
	var errTooLarge files.ErrFileIsTooLarge
	require.ErrorAs(t, err, &errTooLarge, "got %v", err)

	buf, err := c.download("8", "small.bin")
	require.NoError(t, err)
	assert.Equal(t, "hello", buf.String())
	reqs := f.requestsTo("/attachments/8/download/small.bin")
	require.Len(t, reqs, 1)
	assert.Equal(t, "key", reqs[0].Header.Get("X-Api-Key"))
}
