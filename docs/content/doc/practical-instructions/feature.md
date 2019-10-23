---
date: "2019-02-12:00:00+02:00"
title: "Add a new api endpoint"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "practical instructions"
---

# Add a new api endpoint/feature

Most of the api endpoints/features of Vikunja are using the [common web handler](https://code.vikunja.io/web).
This is a library created by Vikunja in an effort to facilitate the creation of REST endpoints.

This works by abstracting the handling of CRUD-Requests, including rights check.

You can learn more about the web handler on [the project's repo](https://code.vikunja.io/web).

### Helper for pagination

Pagination limits can be calculated with a helper function, `getLimitFromPageIndex(pageIndex)` 
(only available in the `models` package) from any page number.
It returns the `limit` (max-length) and `offset` parameters needed for SQL-Queries.

You can feed this function directly into xorm's `Limit`-Function like so:

{{< highlight golang >}}
lists := []List{}
err := x.Limit(getLimitFromPageIndex(pageIndex, itemsPerPage)).Find(&lists)
{{< /highlight >}}

// TODO: Add a full example from start to finish, like a tutorial on how to create a new endpoint?
