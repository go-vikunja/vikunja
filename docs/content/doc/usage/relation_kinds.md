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

| Code | Description |
|------|-------------|
| subtask | Task is a subtask of the other task. This is the opposite of `parenttask`. |
| parenttask | Task is a parent task of the other task. This is the opposite of `subtask`. |
| related | Both tasks are related to each other. How is not more specified. |
| duplicateof | Task is a duplicate of the other task. This is the opposite of `duplicates`. |
| duplicates | Task duplicates the other task. This is the opposite of `duplicateof`. |
| blocking | Task is blocking the other task. This is the opposite of `blocked`. |
| blocked | Task is blocked by the other task. This is the opposite of `blocking`. |
| precedes | Task precedes the other task. This is the opposite of `follows`. |
| follows | Task follows the other task. This is the opposite of `precedes`. |
| copiedfrom | Task is copied from the other task. This is the opposite of `copiedto`. |
| copiedto | Task is copied to the other task. This is the opposite of `copiedfrom`. |
