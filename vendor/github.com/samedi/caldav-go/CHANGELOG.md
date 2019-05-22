# CHANGELOG

v3.0.0
-----------
2017-08-01  Daniel Ferraz  <d.ferrazm@gmail.com>

Main change:

Add two ways to get resources from the storage: shallow or not.

`data.GetShallowResource`: means that, if it's a collection resource, it will not include its child VEVENTs in the ICS data.
This is used throughout the palces where children don't matter.

`data.GetResource`: means that the child VEVENTs will be included in the returned ICS content data for collection resources.
This is used when sending a GET request to fetch a specific resource and expecting its full ICS data in response.

Other changes:

* Removed the need to pass the useless `writer http.ResponseWriter` parameter when calling the `caldav.HandleRequest` function.
* Added a `caldav.HandleRequestWithStorage` function that makes it easy to pass a custom storage to be used and handle the request with a single function call.


v2.0.0
-----------
2017-05-10  Daniel Ferraz  <d.ferrazm@gmail.com>

All commits squashed and LICENSE updated to release as OSS in github.
Feature-wise it remains the same.


v1.0.1
-----------
2017-01-25  Daniel Ferraz  <d.ferrazm@gmail.com>

Escape the contents in `<calendar-data>` and `<displayname>` in the `multistatus` XML responses. Fixing possible bugs
related to having special characters (e.g. &) in the XML multistatus responses that would possibly break the encoding.

v1.0.0
-----------
2017-01-18  Daniel Ferraz  <d.ferrazm@gmail.com>

Main feature:

* Handles the `Prefer` header on PROPFIND and REPORT requests (defined in this [draft/proposal](https://tools.ietf.org/html/draft-murchison-webdav-prefer-05)). Useful to shrink down possible big and verbose responses when the client demands. Ex: current iOS calendar client uses this feature on its PROPFIND requests.

Other changes:

* Added the `handlers.Response` to allow clients of the lib to interact with the generated response before being written/sent back to the client.
* Added `GetResourcesByFilters` to the storage interface to allow filtering of resources in the storage level. Useful to provide an already filtered and smaller resource collection to a the REPORT handler when dealing with a filtered REPORT request.
* Added `GetResourcesByList` to the storage interface to fetch a set a of resources based on a set of paths. Useful to provide, in one call, the correct resource collection to the REPORT handler when dealing with a REPORT request for specific `hrefs`.
* Remove useless `IsResourcePresent` from the storage interface.


v0.1.0
-----------
2016-09-23  Daniel Ferraz  <d.ferrazm@gmail.com>

This version implements:

* Allow: "GET, HEAD, PUT, DELETE, OPTIONS, PROPFIND, REPORT"
* DAV: "1, 3, calendar-access"
* Also only handles the following components: `VCALENDAR`, `VEVENT`

Currently unsupported:

* Components `VTODO`, `VJOURNAL`, `VFREEBUSY`
* `VEVENT` recurrences
* Resource locking
* User authentication
