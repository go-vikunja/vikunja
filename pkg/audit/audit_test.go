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

package audit_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/audit"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	config.InitDefaultConfig()
	keyvalue.InitStorage() // license.SetForTests persists state through keyvalue
	os.Exit(m.Run())
}

// One event type per test so each topic has exactly the listeners the test registered.
type pipelineEvent struct {
	TaskID int64 `json:"task_id"`
	DoerID int64 `json:"doer_id"`
}

func (e *pipelineEvent) Name() string { return "test.audit.pipeline" }

type licenseGateEvent struct {
	Marker string `json:"marker"`
}

func (e *licenseGateEvent) Name() string { return "test.audit.licensegate" }

type rotationEvent struct {
	Filler string `json:"filler"`
}

func (e *rotationEvent) Name() string { return "test.audit.rotation" }

// otherListener is a second, non-audit listener on the same topic.
type otherListener struct {
	called chan struct{}
}

func (l *otherListener) Handle(_ *message.Message) error {
	select {
	case l.called <- struct{}{}:
	default:
	}
	return nil
}

func (l *otherListener) Name() string { return "other" }

var (
	registerTestEventsOnce sync.Once
	other                  = &otherListener{called: make(chan struct{}, 16)}
)

// The listener registry is global and watermill rejects duplicate handler
// names, so register once per process (relevant for -count > 1).
func registerTestEvents() {
	registerTestEventsOnce.Do(func() {
		audit.RegisterEventForAudit(func(e *pipelineEvent) *audit.Entry {
			return &audit.Entry{
				Action: "task.created",
				Actor:  audit.UserActor(e.DoerID),
				Target: audit.TaskTarget(e.TaskID),
			}
		})
		events.RegisterListener((&pipelineEvent{}).Name(), other)

		audit.RegisterEventForAudit(func(e *licenseGateEvent) *audit.Entry {
			return &audit.Entry{
				Action:   "task.created",
				Actor:    audit.SystemActor(),
				Target:   audit.TaskTarget(1),
				Metadata: map[string]any{"marker": e.Marker},
			}
		})

		audit.RegisterEventForAudit(func(e *rotationEvent) *audit.Entry {
			return &audit.Entry{
				Action:   "task.created",
				Actor:    audit.SystemActor(),
				Target:   audit.TaskTarget(1),
				Metadata: map[string]any{"filler": e.Filler},
			}
		})
	})
}

func setupAuditFile(t *testing.T) string {
	t.Helper()
	logfile := filepath.Join(t.TempDir(), "audit.log")
	config.AuditLogfile.Set(logfile)
	require.NoError(t, audit.Init())
	t.Cleanup(audit.Close)
	return logfile
}

func startEventRouter(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ready, err := events.InitEventsForTesting(ctx)
	require.NoError(t, err)
	<-ready
}

func waitForLines(t *testing.T, logfile string) []string {
	t.Helper()
	var lines []string
	require.Eventually(t, func() bool {
		content, err := os.ReadFile(logfile)
		if err != nil {
			return false
		}
		lines = strings.Split(strings.TrimSpace(string(content)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			lines = nil
		}
		return len(lines) >= 1
	}, 5*time.Second, 10*time.Millisecond, "expected at least one audit log line")
	return lines
}

func TestAuditPipeline(t *testing.T) {
	logfile := setupAuditFile(t)
	license.SetForTests([]license.Feature{license.FeatureAuditLogs})
	t.Cleanup(license.ResetForTests)

	registerTestEvents()
	startEventRouter(t)

	ctx := events.WithRequestMeta(context.Background(), &events.RequestMeta{
		IP:        "192.0.2.42",
		UserAgent: "test-agent/1.0",
		RequestID: "req-123",
	})
	require.NoError(t, events.DispatchWithContext(ctx, &pipelineEvent{TaskID: 99, DoerID: 7}))

	waitForLines(t, logfile)
	select {
	case <-other.called:
	case <-time.After(5 * time.Second):
		t.Fatal("other listener on the same topic was not called")
	}
	// A topic with multiple listeners must produce exactly one audit entry.
	events.WaitForPendingHandlers()
	lines := waitForLines(t, logfile)
	require.Len(t, lines, 1)

	var entry audit.Entry
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &entry))
	assert.NotEmpty(t, entry.EventID)
	assert.False(t, entry.Timestamp.IsZero())
	assert.Equal(t, "task.created", entry.Action)
	assert.Equal(t, audit.UserActor(7), entry.Actor)
	assert.Equal(t, audit.TaskTarget(99), entry.Target)
	assert.Equal(t, audit.OutcomeSuccess, entry.Outcome)
	assert.Equal(t, "192.0.2.42", entry.Source.IP)
	assert.Equal(t, "test-agent/1.0", entry.Source.UserAgent)
	assert.Equal(t, audit.SourceHTTP, entry.Source.Type)
	assert.Equal(t, "req-123", entry.RequestID)
}

func TestAuditLicenseGating(t *testing.T) {
	logfile := setupAuditFile(t)

	registerTestEvents()
	startEventRouter(t)

	// Without the licensed feature nothing must be written. The license check
	// happens per event at handle time, so give the async handler a settle
	// window before flipping the license back on.
	license.ResetForTests()
	require.NoError(t, events.Dispatch(&licenseGateEvent{Marker: "unlicensed"}))
	require.Never(t, func() bool {
		content, err := os.ReadFile(logfile)
		return err == nil && len(content) > 0
	}, 500*time.Millisecond, 10*time.Millisecond, "unlicensed event must not be written")
	events.WaitForPendingHandlers()

	license.SetForTests([]license.Feature{license.FeatureAuditLogs})
	t.Cleanup(license.ResetForTests)
	require.NoError(t, events.Dispatch(&licenseGateEvent{Marker: "licensed"}))

	lines := waitForLines(t, logfile)
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], `"marker":"licensed"`)
	assert.NotContains(t, lines[0], "unlicensed")
	assert.Contains(t, lines[0], `"type":"system"`)
}

func TestAuditRotation(t *testing.T) {
	logfile := setupAuditFile(t)
	license.SetForTests([]license.Feature{license.FeatureAuditLogs})
	t.Cleanup(license.ResetForTests)

	registerTestEvents()
	startEventRouter(t)

	// Default max size is 100MB and config values are MB-granular, so two
	// entries of ~600KB cross the limit with maxsizemb set to 1.
	config.AuditRotationMaxSizeMB.Set("1")
	t.Cleanup(func() { config.AuditRotationMaxSizeMB.Set("100") })
	require.NoError(t, audit.Init())

	filler := strings.Repeat("x", 600*1024)
	require.NoError(t, events.Dispatch(&rotationEvent{Filler: filler}))
	waitForLines(t, logfile)
	require.NoError(t, events.Dispatch(&rotationEvent{Filler: filler}))
	waitForLines(t, logfile)

	require.Eventually(t, func() bool {
		rotated, err := filepath.Glob(strings.TrimSuffix(logfile, ".log") + "-*.log")
		return err == nil && len(rotated) == 1
	}, 5*time.Second, 10*time.Millisecond, "expected one rotated audit log file")
}

func TestWriteAuditEventNotInitialized(t *testing.T) {
	audit.Close()
	err := audit.WriteAuditEvent(&audit.Entry{Action: "task.created"})
	require.Error(t, err)
}
