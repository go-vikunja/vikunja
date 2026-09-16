---
name: prepare-worktree
description: Set up an isolated git worktree for implementing a plan, via `mage dev:prepare-worktree`. Use when asked to "prepare a worktree", "set up a worktree for this plan", or to implement something in a separate workspace rather than on the current branch.
---

# Preparing a Worktree for Implementation

```bash
mage dev:prepare-worktree <name> <plan-path>
```

**Arguments:**
- `<name>` - Required. Becomes both the folder name and branch name. Use conventions like `fix-<description>` for bug fixes or `feat-<description>` for new features.
- `<plan-path>` - Required. Path to a plan file (relative to repo root) that will be copied to the new worktree's `plans/` directory. Pass `""` to skip copying a plan.

This will initialize a new worktree in the parent directory and copy some files over — `config.yml` with an updated rootpath, and the frontend gets initialized.

**Example:**
```bash
# Create worktree for a bug fix with a plan
mage dev:prepare-worktree fix-position-healing plans/fix-position-healing.md

# Create worktree for a new feature without a plan
mage dev:prepare-worktree feat-dark-mode ""
```

**Result:**
```
parent-directory/
├── main/                    # Original workspace
├── fix-position-healing/    # New worktree
│   ├── config.yml           # With updated rootpath
│   └── plans/
│       └── fix-position-healing.md
└── ...
```

After creation, tell the user where they can find the new worktree.
