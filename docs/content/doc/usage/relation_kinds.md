---
date: "2019-09-25:00:00+02:00"
title: "Task Relation kinds"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "usage"
---

# Available task relation kinds

| Code | Description | Opposite of |
|------|-------------|-------------|
| `subtask` | Task is a subtask of the other task.       | `parenttask` |
| `parenttask` | Task is a parent task of the other task. | `subtask` |
| `related` | Both tasks are related to each other.<br  /> How is not more specified. | ⸺ |
| `duplicateof` | Task is a duplicate of the other task. | `duplicates` |
| `duplicates` | Task duplicates the other task. | `duplicateof` |
| `blocking` | Task is blocking the other task. | `blocked` |
| `blocked` | Task is blocked by the other task. | `blocking` |
| `precedes` | Task precedes the other task. | `follows` |
| `follows` | Task follows the other task. | `precedes` |
| `copiedfrom` | Task is copied from the other task. | `copiedto` |
| `copiedto` | Task is copied to the other task. | `copiedfrom` |
