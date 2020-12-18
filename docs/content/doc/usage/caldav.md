---
date: "2019-05-12:00:00+01:00"
title: "Caldav"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "usage"
---

# Caldav

> **Warning:** The caldav integration is in an early alpha stage and has bugs. 
> It works well with some clients while having issues with others.
> If you encounter issues, please [report them](https://code.vikunja.io/api/issues/new?body=[caldav]) 

Vikunja supports managing tasks via the [caldav VTODO](https://tools.ietf.org/html/rfc5545#section-3.6.2) extension.

{{< table_of_contents >}}

## URLs

All urls are located under the `/dav` subspace.

Urls are:

* `/principals/<username>/`: Returns urls for list discovery. *Use this url to initially make connections to new clients.*
* `/lists/`: Used to manage lists
* `/lists/<List ID>/`: Used to manage a single list
* `/lists/<List ID>/<Task UID>`: Used to manage a task on a list

## Supported properties

Vikunja currently supports the following properties:

* `UID`
* `SUMMARY`
* `DESCRIPTION`
* `PRIORITY`
* `COMPLETED`
* `DUE`
* `DTSTART`
* `DURATION`
* `ORGANIZER`
* `RELATED-TO`
* `CREATED`
* `DTSTAMP`
* `LAST-MODIFIED`

Vikunja **currently does not** support these properties:

* `ATTACH`
* `CATEGORIES`
* `CLASS`
* `COMMENT`
* `GEO`
* `LOCATION`
* `PERCENT-COMPLETE`
* `RESOURCES`
* `STATUS`
* `CONTACT`
* `RECURRENCE-ID`
* `URL`
* Recurrence
* `SEQUENCE`

## Tested Clients

### Working

* [Evolution](https://wiki.gnome.org/Apps/Evolution/)
* [OpenTasks](https://opentasks.app/) + [DAVx⁵](https://www.davx5.com/)
* [Tasks (Android)](https://tasks.org/)

### Not working

* [Thunderbird (68)](https://www.thunderbird.net/)

## Dev logs

The whole thing is not optimized at all and probably pretty inefficent.

Request body and headers are logged if the debug output is enabled.

```
Creating a new task:
PUT /dav/lists/1/cd4dd0e1b3c19cc9d787829b6e08be536e3df3a4.ics

Body:

BEGIN:VCALENDAR
CALSCALE:GREGORIAN
PRODID:-//Ximian//NONSGML Evolution Calendar//EN
VERSION:2.0
BEGIN:VTODO
UID:cd4dd0e1b3c19cc9d787829b6e08be536e3df3a4
DTSTAMP:20190508T134538Z
SUMMARY:test2000
PRIORITY:0
CLASS:PUBLIC
CREATED:20190508T134710Z
LAST-MODIFIED:20190508T134710Z
END:VTODO
END:VCALENDAR


Marking a task as done:

BEGIN:VCALENDAR
CALSCALE:GREGORIAN
PRODID:-//Ximian//NONSGML Evolution Calendar//EN
VERSION:2.0
BEGIN:VTODO
UID:3ada92f28b4ceda38562ebf047c6ff05400d4c572352a
DTSTAMP:20190511T183631
DTSTART:19700101T000000
DTEND:19700101T000000
SUMMARY:sdgs
ORGANIZER;CN=:user
CREATED:20190511T183631
PRIORITY:0
LAST-MODIFIED:20190512T193428Z
COMPLETED:20190512T193428Z
PERCENT-COMPLETE:100
STATUS:COMPLETED
END:VTODO
END:VCALENDAR

Requests from the app:::

[CALDAV] Request Body: <?xml version="1.0" encoding="UTF-8" ?><propfind xmlns="DAV:" xmlns:CAL="urn:ietf:params:xml:ns:caldav" xmlns:CARD="urn:ietf:params:xml:ns:carddav"><prop><current-user-principal /></prop></propfind>
[CALDAV] GetResources: rpath: /dav/
2019-05-18T23:25:49.971140654+02:00: WEB 	▶ 192.168.1.134  PROPFIND 207 /dav/ 1.021705664s - okhttp/3.12.2

[CALDAV] Request Body: <?xml version="1.0" encoding="UTF-8" ?><propfind xmlns="DAV:" xmlns:CAL="urn:ietf:params:xml:ns:caldav" xmlns:CARD="urn:ietf:params:xml:ns:carddav"><prop><CAL:calendar-home-set /></prop></propfind>
[CALDAV] GetResources: rpath: /dav/
2019-05-18T23:25:52.166996113+02:00: WEB 	▶ 192.168.1.134  PROPFIND 207 /dav/ 1.042834467s - okhttp/3.12.2

And then it just stops.
... and complains about not being able to find the home set
... without even requesting it...


```