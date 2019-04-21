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
| 2001 | 400 | ID cannot be empty or 0. |
| 2002 | 400 | Some of the request data was invalid. The response contains an aditional array with all invalid fields. |
| 3001 | 404 | The list does not exist. |
| 3004 | 403 | The user needs to have read permissions on that list to perform that action. |
| 3005 | 400 | The list title cannot be empty. |
| 4001 | 400 | The list task text cannot be empty. |
| 4002 | 404 | The list task does not exist. |
| 4003 | 403 | All bulk editing tasks must belong to the same list. |
| 4004 | 403 | Need at least one task when bulk editing tasks. |
| 4005 | 403 | The user does not have the right to see the task. |
| 5001 | 404 | The namspace does not exist. | 
| 5003 | 403 | The user does not have access to the specified namespace. |
| 5006 | 400 | The namespace name cannot be empty. |
| 5009 | 403 | The user needs to have namespace read access to perform that action. |
| 5010 | 403 | This team does not have access to that namespace. |
| 5011 | 409 | This user has already access to that namespace. |
| 6001 | 400 | The team name cannot be emtpy. |
| 6002 | 404 | The team does not exist. |
| 6004 | 409 | The team already has access to that namespace or list. |
| 6005 | 409 | The user is already a member of that team. |
| 6006 | 400 | Cannot delete the last team member. |
| 6007 | 403 | The team does not have access to the list to perform that action. |
| 7002 | 409 | The user already has access to that list. |
| 7003 | 403 | The user does not have access to that list. |
| 8001 | 403 | This label already exists on that task. |
| 8002 | 404 | The label does not exist. |
| 8003 | 403 | The user does not have access to this label. |
| 9001 | 403 | The right is invalid. | 