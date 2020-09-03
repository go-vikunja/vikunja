---
date: "2019-02-12:00:00+02:00"
title: "Rights"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "usage"
---

# List and namespace rights for teams and users

Whenever you share a list or namespace with a user or team, you can specify a `rights` parameter. 
This parameter controls the rights that team or user is going to have (or has, if you request the current sharing status).

Rights are being specified using integers.

The following values are possible:

| Right (int) | Meaning |
|-------------|---------|
| 0 (Default) | Read only. Anything which is shared with this right cannot be edited. |
| 1 | Read and write. Namespaces or lists shared with this right can be read and written to by the team or user. |
| 2 | Admin. Can do anything like read and write, but can additionally manage sharing options. |

## Team admins

When adding or querying a team, every member has an additional boolean value stating if it is admin or not.
A team admin can also add and remove team members and also change whether a user in the team is admin or not.