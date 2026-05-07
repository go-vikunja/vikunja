package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"code.vikunja.io/veans/internal/client"
)

// TestClaim_AssignsBotMovesToInProgressTagsBranch exercises the full claim
// flow: assignment, bucket transition, and branch label application.
func TestClaim_AssignsBotMovesToInProgressTagsBranch(t *testing.T) {
	ws, h := provisionWorkspace(t)

	out, _, code := h.Run(t, ws, "--json", "create", "claim me")
	if code != 0 {
		t.Fatalf("create exit %d\n%s", code, out)
	}
	var created client.Task
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatal(err)
	}
	id := fmt.Sprintf("%d", created.Index)

	// Switch the workspace's git branch so claim has something to label with.
	gitInWorkspace(t, ws, "checkout", "-q", "-b", "feat-claim-test")

	_, errOut, code := h.Run(t, ws, "claim", id)
	if code != 0 {
		t.Fatalf("claim exit %d\n%s", code, errOut)
	}

	server := h.GetTask(t, created.ID)

	// Verify bucket transition by reading the workspace's .veans.yml — the
	// bot's expected In Progress bucket is stored there.
	cfg := loadConfig(t, ws)
	if server.BucketID != cfg.Buckets.InProgress {
		t.Fatalf("task not in In Progress bucket: got %d, want %d", server.BucketID, cfg.Buckets.InProgress)
	}

	// Bot assigned.
	assigned := false
	for _, a := range server.Assignees {
		if a != nil && a.ID == cfg.Bot.UserID {
			assigned = true
			break
		}
	}
	if !assigned {
		t.Fatalf("bot %d not in assignees: %+v", cfg.Bot.UserID, server.Assignees)
	}

	// Branch label attached.
	branchLabel := "veans:branch:feat-claim-test"
	if !taskHasLabelTitle(server, branchLabel) {
		t.Fatalf("expected label %q on task; got %+v", branchLabel, server.Labels)
	}
}
