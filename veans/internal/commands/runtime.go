package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/credentials"
	"code.vikunja.io/veans/internal/output"
)

// runtime bundles the artifacts every non-init command needs: parsed config,
// credential store, and an authed HTTP client. Loaded lazily by loadRuntime
// at command start.
type runtime struct {
	cfg    *config.Config
	store  credentials.Store
	client *client.Client
}

func loadRuntime() (*runtime, error) {
	path, err := config.Find("")
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, output.Wrap(output.CodeNotConfigured, err,
				"no .veans.yml found — run `veans init` in your repo first")
		}
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	store := credentials.Default()
	tok, err := store.Get(cfg.Server, cfg.Bot.Username)
	if err != nil {
		return nil, output.Wrap(output.CodeAuth, err,
			"no token for %s on %s — run `veans login` to mint a fresh one",
			cfg.Bot.Username, cfg.Server)
	}
	return &runtime{
		cfg:    cfg,
		store:  store,
		client: client.New(cfg.Server, tok),
	}, nil
}

// resolveTaskID accepts PROJ-NN, #NN, or a bare integer and returns the
// numeric task ID. The project identifier from .veans.yml is used to verify
// the prefix matches; mismatches error out so an agent can't accidentally
// poke a task in the wrong project.
func (r *runtime) resolveTaskID(ctx context.Context, raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, output.New(output.CodeValidation, "empty task ID")
	}

	// #NN form
	if strings.HasPrefix(raw, "#") {
		n, err := strconv.ParseInt(raw[1:], 10, 64)
		if err != nil {
			return 0, output.Wrap(output.CodeValidation, err, "invalid task ID %q", raw)
		}
		return r.lookupByIndex(ctx, n)
	}

	// Bare integer — treat as task index in the configured project.
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return r.lookupByIndex(ctx, n)
	}

	// PROJ-NN form
	idx := strings.LastIndex(raw, "-")
	if idx > 0 && idx < len(raw)-1 {
		prefix := raw[:idx]
		num := raw[idx+1:]
		if r.cfg.ProjectIdentifier != "" && !strings.EqualFold(prefix, r.cfg.ProjectIdentifier) {
			return 0, output.New(output.CodeValidation,
				"task %q has identifier %q, but this repo's .veans.yml uses %q",
				raw, prefix, r.cfg.ProjectIdentifier)
		}
		n, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return 0, output.Wrap(output.CodeValidation, err, "invalid task ID %q", raw)
		}
		return r.lookupByIndex(ctx, n)
	}

	return 0, output.New(output.CodeValidation, "invalid task ID %q (expected PROJ-NN, #NN, or NN)", raw)
}

// lookupByIndex resolves a 1-based per-project task index (the NN in
// PROJ-NN / #NN) to a numeric task ID by listing the project's tasks and
// matching on Index. The cost is one paged GET; we tolerate it because
// resolving by index without a dedicated endpoint is the only stable path.
func (r *runtime) lookupByIndex(ctx context.Context, index int64) (int64, error) {
	tasks, err := r.client.ListProjectTasks(ctx, r.cfg.ProjectID, &client.TaskListOptions{
		Filter: fmt.Sprintf("index = %d", index),
	})
	if err != nil {
		return 0, err
	}
	for _, t := range tasks {
		if t.Index == index {
			return t.ID, nil
		}
	}
	return 0, output.New(output.CodeNotFound, "task %s not found in project %d",
		r.cfg.FormatTaskID(index), r.cfg.ProjectID)
}
