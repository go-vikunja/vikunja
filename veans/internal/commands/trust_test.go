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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	"code.vikunja.io/veans/internal/auth"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/credentials"
	"code.vikunja.io/veans/internal/trust"
)

// envLikeStore hands out a token for any server, like VEANS_TOKEN does.
type envLikeStore struct{}

func (envLikeStore) Get(_, _ string) (string, error) { return "env-token", nil }
func (envLikeStore) Set(_, _, _ string) error        { return errors.New("read-only") }
func (envLikeStore) Name() string                    { return "env-like" }

var _ credentials.Store = envLikeStore{}

func countingServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/api/v2/projects/1":
			_, _ = w.Write([]byte(`{"id":1,"title":"Remote Title"}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func untrustedGate(t *testing.T) *trust.Gate {
	t.Helper()
	t.Setenv("VEANS_SERVER", "")
	return &trust.Gate{Store: trust.NewStore(filepath.Join(t.TempDir(), "trusted.json"))}
}

func repoConfig(t *testing.T, server string) *config.Config {
	t.Helper()
	cfg := &config.Config{Server: server, ProjectID: 1, Bot: config.Bot{Username: "bot-x"}}
	if err := cfg.SaveAs(filepath.Join(t.TempDir(), config.Filename)); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestNewRuntime_UntrustedSendsNothing(t *testing.T) {
	srv, hits := countingServer(t)
	// No frontend_url, so a trusted runtime would backfill via GET /info.
	cfg := repoConfig(t, srv.URL)

	_, err := newRuntime(context.Background(), cfg, untrustedGate(t), envLikeStore{})
	if !errors.Is(err, trust.ErrUntrusted) {
		t.Fatalf("got %v, want ErrUntrusted", err)
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("untrusted server received %d requests", n)
	}
}

func TestNewRuntime_Trusted(t *testing.T) {
	srv, hits := countingServer(t)
	cfg := repoConfig(t, srv.URL)
	gate := untrustedGate(t)
	if err := gate.Record(srv.URL); err != nil {
		t.Fatal(err)
	}

	rt, err := newRuntime(context.Background(), cfg, gate, envLikeStore{})
	if err != nil {
		t.Fatal(err)
	}
	if rt.client.Token != "env-token" {
		t.Fatalf("token = %q", rt.client.Token)
	}
	if hits.Load() == 0 {
		t.Fatal("expected the frontend_url backfill request")
	}
}

func TestNewRuntime_UserinfoServerRejected(t *testing.T) {
	srv, hits := countingServer(t)
	gate := untrustedGate(t)
	if err := gate.Record(srv.URL); err != nil {
		t.Fatal(err)
	}
	cfg := repoConfig(t, "http://trusted.example@"+srv.Listener.Addr().String())

	if _, err := newRuntime(context.Background(), cfg, gate, envLikeStore{}); err == nil {
		t.Fatal("userinfo server passed the gate")
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("server received %d requests", n)
	}
}

func TestLogin_UntrustedSendsNoCredentials(t *testing.T) {
	srv, hits := countingServer(t)
	cfg := repoConfig(t, srv.URL)

	err := runLogin(context.Background(), cfg, untrustedGate(t), auth.LoginOptions{
		UsePassword: true,
		Username:    "victim",
		Password:    "hunter2",
	}, auth.NewStdPrompter())
	if !errors.Is(err, trust.ErrUntrusted) {
		t.Fatalf("got %v, want ErrUntrusted", err)
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("untrusted server received %d requests", n)
	}
}

func TestPrimeContext_UntrustedRendersWithoutRemoteData(t *testing.T) {
	srv, hits := countingServer(t)
	cfg := repoConfig(t, srv.URL)

	data := buildPrimeContext(context.Background(), cfg, untrustedGate(t), envLikeStore{})
	if data.ProjectTitle != "(unknown)" {
		t.Fatalf("title = %q, want (unknown)", data.ProjectTitle)
	}
	want, err := trust.Origin(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if data.UntrustedServer != want {
		t.Fatalf("UntrustedServer = %q, want %q", data.UntrustedServer, want)
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("untrusted server received %d requests", n)
	}
}

func TestPrimeContext_TrustedFetchesTitle(t *testing.T) {
	srv, _ := countingServer(t)
	cfg := repoConfig(t, srv.URL)
	gate := untrustedGate(t)
	if err := gate.Record(srv.URL); err != nil {
		t.Fatal(err)
	}

	data := buildPrimeContext(context.Background(), cfg, gate, envLikeStore{})
	if data.ProjectTitle != "Remote Title" {
		t.Fatalf("title = %q", data.ProjectTitle)
	}
	if data.UntrustedServer != "" {
		t.Fatalf("unexpected hint %q", data.UntrustedServer)
	}
}
