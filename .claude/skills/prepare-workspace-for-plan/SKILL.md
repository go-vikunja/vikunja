---
name: prepare-workspace-for-plan
description: Use when you have a plan file ready and need to create an isolated git worktree for implementation - creates worktree in parent directory following project conventions and moves the plan file
---

# Prepare Workspace for Plan

Use this skill when you have created or refined a plan and need to set up an isolated workspace for implementation.

## When to Use

- After creating/finalizing a plan in the `plans/` directory
- Before starting implementation of a multi-phase plan
- When you need an isolated branch for a feature or fix

## Prerequisites

- A plan file exists in the current workspace's `plans/` directory
- You are in a git repository that supports worktrees
- The parent directory is the standard location for worktrees (e.g., `/path/to/vikunja/`)

## Steps

### 1. Determine Workspace Name

Choose a name following the project convention:
- `fix-<description>` for bug fixes
- `feat-<description>` for new features

The name should be kebab-case and descriptive but concise.

### 2. Create the Git Worktree

```bash
# From the current workspace (e.g., main/)
git worktree add ../<workspace-name> -b <branch-name>
```

The branch name should match the workspace name.

### 3. Create Plans Directory and Move Plan

```bash
mkdir -p ../<workspace-name>/plans
mv plans/<plan-file>.md ../<workspace-name>/plans/
```

### 4. Verify Structure

```bash
ls -la ../<workspace-name>/plans/
```

## Example

```bash
# Create worktree for position healing fix
git worktree add ../fix-position-healing -b fix-position-healing

# Move the plan
mkdir -p ../fix-position-healing/plans
mv plans/positioning-fixes-detection.md ../fix-position-healing/plans/
```

## Result

After completion, you'll have:
```
parent-directory/
├── main/                    # Original workspace
├── <new-workspace>/         # New worktree
│   └── plans/
│       └── <plan-file>.md   # Your plan
└── ...                      # Other existing worktrees
```

## Notes

- The new worktree shares git history with main but has its own working directory
- Changes in the new worktree won't affect main until merged
- Plans are not committed to git (see `.gitignore`)
- Remember to switch to the new workspace directory to begin implementation
