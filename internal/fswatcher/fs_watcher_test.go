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

package fswatcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	waitForDebounce            = debounce * 120 / 100 // 120% the debounce wait
	waitForShutdownPropagation = 100 * time.Millisecond
)

// sleep is like [time.Sleep] but returns early if the test has completed
func sleep(tb testing.TB, duration time.Duration) {
	tb.Helper()
	select {
	case <-tb.Context().Done():
		tb.Fatal(tb.Context().Err())
	case <-time.After(duration):
	}
}

func touchFile(tb testing.TB, name string) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		tb.Error(err)
		return
	}
	f, err := os.Create(name)
	if err != nil {
		tb.Error(err)
		return
	}
	if err := f.Close(); err != nil {
		tb.Error(err)
		return
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	const watchFileExtension = ".ext"

	type FileUpdate struct {
		Wait       time.Duration // waits this long after performing file operations
		CreatePath string        // creates file at this path, including any intermediate directories
		DeletePath string        // recursively deletes this path
	}
	const quickUpdate = 100 * time.Millisecond
	for _, tc := range []struct {
		description    string
		updates        []FileUpdate
		expectTriggers int64
	}{
		{
			description:    "no updates tears down cleanly",
			updates:        nil,
			expectTriggers: 0,
		},
		{
			description: "first update does not trigger right away",
			updates: []FileUpdate{
				{CreatePath: "foo.ext", Wait: quickUpdate},
			},
			expectTriggers: 0,
		},
		{
			description: "first update",
			updates: []FileUpdate{
				{CreatePath: "foo.ext", Wait: waitForDebounce},
			},
			expectTriggers: 1,
		},
		{
			description: "second update",
			updates: []FileUpdate{
				{CreatePath: "foo.ext", Wait: waitForDebounce},
				{CreatePath: "foo.ext", Wait: waitForDebounce},
			},
			expectTriggers: 2,
		},
		{
			description: "multiple fast updates triggers once",
			updates: []FileUpdate{
				{CreatePath: "foo.ext", Wait: quickUpdate},
				{CreatePath: "foo.ext", Wait: quickUpdate},
				{CreatePath: "foo.ext", Wait: quickUpdate},
				{CreatePath: "foo.ext", Wait: waitForDebounce},
			},
			expectTriggers: 1,
		},
		{
			description: "inner directory also watched",
			updates: []FileUpdate{
				{CreatePath: "foo/bar/baz.ext", Wait: quickUpdate},
				{CreatePath: "foo/biff/boo.ext", Wait: waitForDebounce},
				{CreatePath: "foo/bar/baz.ext", Wait: quickUpdate},
				{CreatePath: "foo/biff/boo.ext", Wait: waitForDebounce},
			},
			expectTriggers: 2,
		},
		{
			description: "file removed",
			updates: []FileUpdate{
				{CreatePath: "foo.ext", Wait: waitForDebounce},
				{DeletePath: "foo.ext", Wait: waitForDebounce},
			},
			expectTriggers: 2,
		},
		{
			description: "directory removed",
			updates: []FileUpdate{
				{CreatePath: "foo/bar.ext", Wait: waitForDebounce},
				{DeletePath: "foo", Wait: waitForDebounce},
			},
			expectTriggers: 2,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(t.Context())
			dir := t.TempDir()

			var triggers atomic.Int64
			watcher, err := newWatcher(ctx, []string{dir}, "*"+watchFileExtension, func() {
				t.Log("triggered")
				triggers.Add(1)
			})
			require.NoError(t, err)

			for _, update := range tc.updates {
				if update.CreatePath != "" {
					t.Log("Creating file:", update.CreatePath)
					touchFile(t, filepath.Join(dir, update.CreatePath))
				}
				if update.DeletePath != "" {
					t.Log("Deleting:", update.DeletePath)
					require.NoError(t, os.RemoveAll(filepath.Join(dir, update.DeletePath)))
				}
				sleep(t, update.Wait)
			}

			cancel()
			const cleanupWait = 5 * time.Second
			assert.Equal(t, context.Canceled, watcher.Wait(cleanupWait))

			assert.Equal(t, tc.expectTriggers, triggers.Load())
		})
	}
}

func TestNewRestarter(t *testing.T) {
	t.Parallel()
	type runErrs struct {
		startCtx, startWatchCtx string
		endCtx, endWatchCtx     string
	}
	ctxErrsChan := make(chan runErrs, 2)

	watchCtx, watchCancel := context.WithCancelCause(t.Context())
	dir := t.TempDir()
	watcher, err := newRestarter(watchCtx, []string{dir}, "*.ext", func(ctx context.Context) error {
		var r runErrs

		r.startCtx = fmt.Sprint(context.Cause(ctx))
		r.startWatchCtx = fmt.Sprint(context.Cause(watchCtx))
		<-ctx.Done()
		r.endCtx = fmt.Sprint(context.Cause(ctx))
		r.endWatchCtx = fmt.Sprint(context.Cause(watchCtx))

		ctxErrsChan <- r
		if watchCtx.Err() != nil {
			close(ctxErrsChan)
		}
		return nil
	})
	require.NoError(t, err)

	touchFile(t, filepath.Join(dir, "foo.ext"))
	sleep(t, waitForDebounce)
	shutdownErr := errors.New("watcher shutdown cause")
	watchCancel(shutdownErr)
	sleep(t, waitForShutdownPropagation)

	const cleanupWait = 5 * time.Second
	assert.Equal(t, shutdownErr, watcher.Wait(cleanupWait))

	var ctxErrs []runErrs
	for r := range ctxErrsChan {
		ctxErrs = append(ctxErrs, r)
	}
	assert.Equal(t, []runErrs{
		{
			startCtx:      "<nil>",
			startWatchCtx: "<nil>",
			endCtx:        "files updated matching pattern: *.ext", // should cancel 'run' when a file changes
			endWatchCtx:   "<nil>",
		},
		{
			startCtx:      "<nil>",
			startWatchCtx: "<nil>",
			endCtx:        shutdownErr.Error(),
			endWatchCtx:   shutdownErr.Error(), // should shutdown 'run' when main context is canceled
		},
	}, ctxErrs)
}

func TestWatch(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	dir := t.TempDir()

	go func() {
		touchFile(t, filepath.Join(dir, "foo.ext"))
		sleep(t, waitForDebounce)
		cancel()
	}()

	var runCount atomic.Int64
	err := Watch(ctx, []string{dir}, "*.ext", func(ctx context.Context) error {
		<-ctx.Done()
		runCount.Add(1)
		return nil
	})
	assert.Equal(t, context.Canceled, err)
	sleep(t, waitForShutdownPropagation)
	assert.EqualValues(t, 2, runCount.Load())
}
