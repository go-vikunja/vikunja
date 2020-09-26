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
| 1001 | 400 | A user with this username already exists. |
| 1002 | 400 | A user with this email address already exists. |
| 1004 | 400 | No username and password specified. |
| 1005 | 404 | The user does not exist. |
| 1006 | 400 | Could not get the user id. |
| 1008 | 412 | No password reset token provided. |
| 1009 | 412 | Invalid password reset token. |
| 1010 | 412 | Invalid email confirm token. |
| 1011 | 412 | Wrong username or password. |
| 1012 | 412 | Email address of the user not confirmed. |
| 1013 | 412 | New password is empty. |
| 1014 | 412 | Old password is empty. |
| 1015 | 412 | Totp is already enabled for this user. |
| 1016 | 412 | Totp is not enabled for this user. |
| 1017 | 412 | The provided Totp passcode is invalid. |
| 1018 | 412 | The provided user avatar provider type setting is invalid. |

## Validation

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 2001 | 400 | ID cannot be empty or 0. |
| 2002 | 400 | Some of the request data was invalid. The response contains an aditional array with all invalid fields. |

## List

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 3001 | 404 | The list does not exist. |
| 3004 | 403 | The user needs to have read permissions on that list to perform that action. |
| 3005 | 400 | The list title cannot be empty. |
| 3006 | 404 | The list share does not exist. |
| 3007 | 400 | A list with this identifier already exists. |
| 3008 | 412 | The list is archived and can therefore only be accessed read only. This is also true for all tasks associated with this list. |

## Task

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 4001 | 400 | The list task text cannot be empty. |
| 4002 | 404 | The list task does not exist. |
| 4003 | 403 | All bulk editing tasks must belong to the same list. |
| 4004 | 403 | Need at least one task when bulk editing tasks. |
| 4005 | 403 | The user does not have the right to see the task. |
| 4006 | 403 | The user tried to set a parent task as the task itself. |
| 4007 | 400 | The user tried to create a task relation with an invalid kind of relation. |
| 4008 | 409 | The user tried to create a task relation which already exists. |
| 4009 | 404 | The task relation does not exist. | 
| 4010 | 400 | Cannot relate a task with itself. |
| 4011 | 404 | The task attachment does not exist. |
| 4012 | 400 | The task attachment is too large. |
| 4013 | 400 | The task sort param is invalid. |
| 4014 | 400 | The task sort order is invalid. |
| 4015 | 404 | The task comment does not exist. |
| 4016 | 403 | Invalid task field. |
| 4017 | 403 | Invalid task filter comparator. |
| 4018 | 403 | Invalid task filter concatinator. |
| 4019 | 403 | Invalid task filter value. |

## Namespace

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 5001 | 404 | The namspace does not exist. | 
| 5003 | 403 | The user does not have access to the specified namespace. |
| 5006 | 400 | The namespace name cannot be empty. |
| 5009 | 403 | The user needs to have namespace read access to perform that action. |
| 5010 | 403 | This team does not have access to that namespace. |
| 5011 | 409 | This user has already access to that namespace. |
| 5012 | 412 | The namespace is archived and can therefore only be accessed read only. |

## Team

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 6001 | 400 | The team name cannot be emtpy. |
| 6002 | 404 | The team does not exist. |
| 6004 | 409 | The team already has access to that namespace or list. |
| 6005 | 409 | The user is already a member of that team. |
| 6006 | 400 | Cannot delete the last team member. |
| 6007 | 403 | The team does not have access to the list to perform that action. |

## User List Access

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 7002 | 409 | The user already has access to that list. |
| 7003 | 403 | The user does not have access to that list. |

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
| 10002 | 400 | The bucket does not belong to that list. | 
| 10003 | 412 | You cannot remove the last bucket on a list. | 
| 10004 | 412 | You cannot add the task to this bucket as it already exceeded the limit of tasks it can hold. | 

## Saved Filters

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 11001 | 404 | The saved filter does not exist. |
| 11002 | 412 | Saved filters are not available for link shares. | 
