---
date: "2019-02-12:00:00+02:00"
title: "Errors"
draft: false
type: "doc"
menu:
sidebar:
parent: "usage"
---

# Errors

This document describes the different errors Vikunja can return.

{{< table_of_contents >}}

## Generic

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 0001 | 403 | Generic forbidden error. |

## User

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 1001      | 400 | A user with this username already exists. |
| 1002      | 400 | A user with this email address already exists. |
| 1004      | 400 | No username and password specified. |
| 1005      | 404 | The user does not exist. |
| 1006      | 400 | Could not get the user id. |
| 1008      | 412 | No password reset token provided. |
| 1009      | 412 | Invalid password reset token. |
| 1010      | 412 | Invalid email confirm token. |
| 1011      | 412 | Wrong username or password. |
| 1012      | 412 | Email address of the user not confirmed. |
| 1013      | 412 | New password is empty. |
| 1014      | 412 | Old password is empty. |
| 1015      | 412 | Totp is already enabled for this user. |
| 1016      | 412 | Totp is not enabled for this user. |
| 1017      | 412 | The provided Totp passcode is invalid. |
| 1018      | 412 | The provided user avatar provider type setting is invalid. |
| 1019      | 412 | No openid email address was provided. |
| 1020      | 412 | This user account is disabled. |
| 1021      | 412 | This account is managed by a third-party authentication provider. |
| 1021      | 412 | The username must not contain spaces. |
| 1022      | 412 | The custom scope set by the OIDC provider is malformed. Please make sure the openid provider sets the data correctly for your scope. Check especially to have set an oidcID. |

## Validation

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 2001 | 400 | ID cannot be empty or 0. |
| 2002 | 400 | Some of the request data was invalid. The response contains an additional array with all invalid fields. |

## Project

| ErrorCode | HTTP Status Code | Description                                                                                                                        |
|-----------|------------------|------------------------------------------------------------------------------------------------------------------------------------|
| 3001      | 404 | The project does not exist.                                                                                                        |
| 3004      | 403 | The user needs to have read permissions on that project to perform that action.                                                    |
| 3005      | 400 | The project title cannot be empty.                                                                                                 |
| 3006      | 404 | The project share does not exist.                                                                                                  |
| 3007      | 400 | A project with this identifier already exists.                                                                                     |
| 3008      | 412 | The project is archived and can therefore only be accessed read only. This is also true for all tasks associated with this project. |
| 3009      | 412 | The project cannot belong to a dynamically generated parent project like "Favorites".                                              |
| 3010      | 412 | This project cannot be a child of itself.                                                                                          |
| 3011      | 412 | This project cannot have a cyclic relationship to a parent project.                                                                |
| 3012      | 412 | This project cannot be deleted because a user has set it as their default project.                                                 |
| 3013      | 412 | This project cannot be archived because a user has set it as their default project.                                                |
| 3014      | 404 | This project view does not exist.                                                                                                  |

## Task

| ErrorCode | HTTP Status Code | Description                                                                |
|-----------|------------------|----------------------------------------------------------------------------|
| 4001      | 400 | The project task text cannot be empty.                                     |
| 4002      | 404 | The project task does not exist.                                           |
| 4003      | 403 | All bulk editing tasks must belong to the same project.                    |
| 4004      | 403 | Need at least one task when bulk editing tasks.                            |
| 4005      | 403 | The user does not have the right to see the task.                          |
| 4006      | 403 | The user tried to set a parent task as the task itself.                    |
| 4007      | 400 | The user tried to create a task relation with an invalid kind of relation. |
| 4008      | 409 | The user tried to create a task relation which already exists.             |
| 4009      | 404 | The task relation does not exist.                                          |
| 4010      | 400 | Cannot relate a task with itself.                                          |
| 4011      | 404 | The task attachment does not exist.                                        |
| 4012      | 400 | The task attachment is too large.                                          |
| 4013      | 400 | The task sort param is invalid.                                            |
| 4014      | 400 | The task sort order is invalid.                                            |
| 4015      | 404 | The task comment does not exist.                                           |
| 4016      | 400 | Invalid task field.                                                        |
| 4017      | 400 | Invalid task filter comparator.                                            |
| 4018      | 400 | Invalid task filter concatinator.                                          |
| 4019      | 400 | Invalid task filter value.                                                 |
| 4020      | 400 | The provided attachment does not belong to that task.                      |
| 4021      | 400 | This user is already assigned to that task.                                |
| 4022      | 400 | The task has a relative reminder which does not specify relative to what.  |
| 4023      | 409 | Tried to create a task relation which would create a cycle.                |
| 4024      | 400 | The provided filter expression is invalid.                                 |
| 4025      | 400 | The reaction kind is invalid.                                              |
| 4026      | 400 | You must provide a project view ID when sorting by position.               |

## Team

| ErrorCode | HTTP Status Code | Description                                                          |
|-----------|------------------|----------------------------------------------------------------------|
| 6001 | 400 | The team name cannot be empty.                                       |
| 6002 | 404 | The team does not exist.                                             |
| 6004 | 409 | The team already has access to that project.                         |
| 6005 | 409 | The user is already a member of that team.                           |
| 6006 | 400 | Cannot delete the last team member.                                  |
| 6007 | 403 | The team does not have access to the project to perform that action. |
| 6008 | 400 | There are no teams found with that team name. |
| 6009 | 400 | There is no oidc team with that team name and oidcId. |
| 6010 | 400 | There are no oidc teams found for the user. |

## User Project Access

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 7002 | 409 | The user already has access to that project. |
| 7003 | 403 | The user does not have access to that project. |

## Label

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 8001 | 403 | This label already exists on that task. |
| 8002 | 404 | The label does not exist. |
| 8003 | 403 | The user does not have access to this label. |

## Right

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 9001 | 403 | The right is invalid. |

## Kanban

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 10001 | 404 | The bucket does not exist. |
| 10002 | 400 | The bucket does not belong to that project. |
| 10003 | 412 | You cannot remove the last bucket on a project. |
| 10004 | 412 | You cannot add the task to this bucket as it already exceeded the limit of tasks it can hold. |
| 10005 | 412 | There can be only one done bucket per project. |

## Saved Filters

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 11001 | 404 | The saved filter does not exist. |
| 11002 | 412 | Saved filters are not available for link shares. |

## Subscriptions

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 12001 | 412 | The subscription entity type is invalid. |
| 12002 | 412 | The user is already subscribed to the entity itself or a parent entity. |

## Link Shares

| ErrorCode | HTTP Status Code | Description                                                                    |
|-----------|------------------|--------------------------------------------------------------------------------|
| 13001 | 412 | This link share requires a password for authentication, but none was provided. |
| 13002 | 403 | The provided link share password is invalid.                                   |
| 13003 | 400 | The provided link share token is invalid.                                      |
