# Null Date Sentinel

Vikunja stores dates in Typesense as Unix timestamps. Because Typesense does not support filtering for missing values, tasks without a due date use a sentinel value of `-1` when indexed. Queries that need to include tasks with no due date can use a boolean OR filter:

```
filter_by=due_date:=-1 || due_date:>=<epoch>
```

The sentinel is chosen from the negative range so it will never conflict with real timestamps which are always positive.

After upgrading, rebuild the index with:

```
vikunja reindex-sentinel
```
