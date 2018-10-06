# Errors

This document describes the different errors Vikunja can return.

| ErrorCode | HTTP Status Code | Description |
|-----------|------------------|-------------|
| 1001 | 400 | A user with this username already exists. |
| 1002 | 400 | A user with this email address already exists. |
| 1004 | 400 | No username and password specified. |
| 1005 | 404 | The user does not exist. |
| 1006 | 400 | Could not get the user id. |
| 1007 | 409 | Cannot delete the last user on the system. |
| 2001 | 400 | ID cannot be empty or 0. |
| 3001 | 404 | The list does not exist. |
| 3004 | 403 | The user needs to have read permissions on that list to perform that action. |
| 3005 | 400 | The list title cannot be empty. |
| 4001 | 400 | The list task text cannot be empty. |
| 4002 | 404 | The list task does not exist. |
| 5001 | 404 | The namspace does not exist. | 
| 5003 | 403 | The user does not have access to the specified namespace. |
| 5006 | 400 | The namespace name cannot be empty. |
| 5009 | 403 | The user needs to have namespace read access to perform that action. |
| 5010 | 403 | This team does not have access to that namespace. |
| 5011 | 409 | This user has already access to that namespace. |
| 6001 | 400 | The team name cannot be emtpy. |
| 6002 | 404 | The team does not exist. |
| 6003 | 400 | The provided team right is invalid. |
| 6004 | 409 | The team already has access to that namespace or list. |
| 6005 | 409 | The user is already a member of that team. |
| 6006 | 400 | Cannot delete the last team member. |
| 6007 | 403 | The team does not have access to the list to perform that action. |
| 7001 | 400 | The user right is invalid. |
| 7002 | 409 | The user already has access to that list. |
| 7003 | 403 | The user does not have access to that list. |
