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
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type watcher struct {
	ctx         context.Context
	filePattern string
	fs          *fsnotify.Watcher
	run         func()
}

// newWatcher returns a [watcher] that calls 'run' for every file update inside one of 'rootDirectories' and the file matches 'filePattern'.
// Stops watching when 'ctx' is canceled.
func newWatcher(ctx context.Context, rootDirectories []string, filePattern string, run func()) (*watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, root := range rootDirectories {
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if err := fsWatcher.Add(path); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	w := &watcher{
		ctx:         ctx,
		filePattern: filePattern,
		fs:          fsWatcher,
		run:         run,
	}
	go w.watch()
	return w, nil
}

const debounce = 2 * time.Second

func (w *watcher) watch() {
	defer w.fs.Close()
	timer := time.NewTimer(0)
	timer.Stop() // avoid firing until update event
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			w.run()
		case <-w.ctx.Done():
			return
		case event := <-w.fs.Events:
			info, infoErr := os.Stat(event.Name)
			isDir := infoErr == nil && info.IsDir()
			if event.Op&fsnotify.Create != 0 {
				if isDir {
					_ = w.fs.Add(event.Name)
				} else {
					_ = w.fs.Add(filepath.Dir(event.Name))
				}
			}
			matched, err := filepath.Match(w.filePattern, filepath.Base(event.Name))
			if (isDir || (err == nil && matched)) && event.Op&(fsnotify.Write|fsnotify.Remove|fsnotify.Create|fsnotify.Rename) != 0 {
				timer.Reset(debounce)
			}
		case err := <-w.fs.Errors:
			var pathErr *os.PathError
			if errors.As(err, &pathErr) && errors.Is(err, os.ErrNotExist) {
				_ = w.fs.Remove(pathErr.Path)
			} else {
				fmt.Fprintln(os.Stderr, "Watch error:", err)
			}
		}
	}
}

// Wait returns when the original context has been canceled. Returns the context's error.
func (w *watcher) Wait(timeout time.Duration) error {
	select {
	case <-w.ctx.Done():
		return context.Cause(w.ctx)
	case <-time.After(timeout):
		return errors.New("timed out waiting for watcher")
	}
}

type restarter struct {
	Context context.Context
	Cancel  context.CancelCauseFunc
}

// newRestarter is like [newWatcher], but automatically restarts 'run' on watch events.
// 'run' is called, its context is canceled when a watch event occurs, then repeats until 'ctx' is canceled.
func newRestarter(ctx context.Context, rootDirectories []string, filePattern string, run func(context.Context) error) (*watcher, error) {
	var runRestarter atomic.Pointer[restarter]
	runRestarter.Store(&restarter{Context: nil, Cancel: func(error) {}})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				runCtx, runCancel := context.WithCancelCause(ctx)
				runRestarter.Store(&restarter{
					Context: runCtx,
					Cancel:  runCancel,
				})
				err := run(runCtx)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				const commandFailedWait = 10 * time.Second
				select {
				case <-runCtx.Done():
				case <-time.After(commandFailedWait):
				}
			}
		}
	}()
	return newWatcher(ctx, rootDirectories, filePattern, func() {
		runRestarter.Load().Cancel(errors.Errorf("files updated matching pattern: %s", filePattern))
	})
}

// Watch runs 'run' for every matching watch event until 'ctx' is canceled.
//
// Specifically 'run' is called, its context is canceled when a watch event occurs, then repeats until 'ctx' is canceled.
func Watch(ctx context.Context, rootDirectories []string, filePattern string, run func(context.Context) error) error {
	watcher, err := newRestarter(ctx, rootDirectories, filePattern, run)
	if err != nil {
		return err
	}
	<-ctx.Done()
	const waitTimeout = 5 * time.Second
	return watcher.Wait(waitTimeout)
}
