# go CalDAV

This is a Go lib that aims to implement the CalDAV specification ([RFC4791]). It allows the quick implementation of a CalDAV server in Go. Basically, it provides the request handlers that will handle the several CalDAV HTTP requests, fetch the appropriate resources, build and return the responses.

### How to install

```
go get github.com/samedi/caldav-go
```

### Dependencies

For dependency management, `glide` is used.

```bash
# install glide (once!)
curl https://glide.sh/get | sh

# install dependencies
glide install
```

### How to use it

The easiest way to quickly implement a CalDAV server is by just using the lib's request handler. Example:

```go
package mycaldav

import (
  "net/http"
  "github.com/samedi/caldav-go"
)

func runServer() {
  http.HandleFunc(PATH, caldav.RequestHandler)
  http.ListenAndServe(PORT, nil)
}
```

With that, all the HTTP requests (GET, PUT, REPORT, PROPFIND, etc) will be handled and responded by the `caldav` handler. In case of any HTTP methods not supported by the lib, a `501 Not Implemented` response will be returned.

In case you want more flexibility to handle the requests, e.g., if you wanted to access the generated response before being sent back to the caller, you could do like:

```go
package mycaldav

import (
  "net/http"
  "github.com/samedi/caldav-go"
)

func runServer() {
  http.HandleFunc(PATH, myHandler)
  http.ListenAndServe(PORT, nil)
}

func myHandler(writer http.ResponseWriter, request *http.Request) {
  response := caldav.HandleRequest(request)
  // ... do something with the response object before writing it back to the client ...
  response.Write(writer)
}
```

### Storage & Resources

The storage is where the CalDAV resources are stored. To interact with that, the caldav lib needs only a type that conforms with the  `data.Storage` interface to operate on top of the storage. Basically, this interface defines all the CRUD functions to work on top of the resources. With that, resources can be stored anywhere: in the filesystem, in the cloud, database, etc. As long as the used storage implements all the required storage interface functions, the caldav lib will work fine.

For example, we could use the following dummy read-only storage implementation, which returns dummy hard-coded resources:

```go
type DummyStorage struct{
  resources map[string]string{
    "/foo": `BEGING:VCALENDAR\nBEGIN:VEVENT\nDTSTART:20160914T170000\nEND:VEVENT\nEND:VCALENDAR`,
    "/bar": `BEGING:VCALENDAR\nBEGIN:VEVENT\nDTSTART:20160915T180000\nEND:VEVENT\nEND:VCALENDAR`,
    "/baz": `BEGING:VCALENDAR\nBEGIN:VEVENT\nDTSTART:20160916T190000\nEND:VEVENT\nEND:VCALENDAR`,
  }
}

func (d *DummyStorage) GetResources(rpath string, withChildren bool) ([]Resource, error) {
  return d.GetResourcesByList([]string{rpath})
}

func (d *DummyStorage) GetResourcesByFilters(rpath string, filters *ResourceFilter) ([]Resource, error) {
  return nil, errors.New("filters are not supported")
}

func (d *DummyStorage) GetResourcesByList(rpaths []string) ([]Resource, error) {
  result := []Resource{}

  for _, rpath := range rpaths {
    resource, found, _ := d.GetResource(rpath)
    if found {
      result = append(result, resource)
    }
  }

  return result, nil
}

func (d *DummyStorage) GetResource(rpath string) (*Resource, bool, error) {
  return d.GetShallowResource(rpath)
}

func (d *DummyStorage) GetShallowResource(rpath string) (*Resource, bool, error) {
  result := []Resource{}
  resContent := d.resources[rpath]

  if resContent != "" {
    resource = NewResource(rpath, DummyResourceAdapter{rpath, resContent})
    return &resource, true, nil
  }

  return nil, false, nil
}

func (d *DummyStorage) CreateResource(rpath, content string) (*Resource, error) {
  return nil, errors.New("creating resources are not supported")
}

func (d *DummyStorage) UpdateResource(rpath, content string) (*Resource, error) {
  return nil, errors.New("updating resources are not supported")
}

func (d *DummyStorage) DeleteResource(rpath string) error {
  return nil, errors.New("deleting resources are not supported")
}
```

Normally, when you provide your own storage implementation, you will need to provide also a custom `data.ResourceAdapter` interface implementation.
The resource adapter deals with the specificities of how resources are stored, which formats and how to deal with them. For example,
for file resources, the resources contents are the content read from the file itself, for resources in the cloud, it could be in JSON needing
some additional processing to parse the content, etc.

In our example here, we could say that the adapter for this case would be:

```go
type DummyResourceAdapter struct {
  resourcePath string
  resourceData string
}

func (a *DummyResourceAdapter) IsCollection() bool {
  return false
}

func (a *DummyResourceAdapter) GetContent() string {
  return a.resourceData
}

func (a *DummyResourceAdapter) GetContentSize() int64 {
  return len(a.GetContent())
}

func (a *DummyResourceAdapter) CalculateEtag() string {
  return hashify(a.GetContent())
}

func (a *DummyResourceAdapter) GetModTime() time.Time {
  return time.Now()
}
```

Note that this adapter implementation is passed over whenever we initialize a new `Resource` instance in the storage implementation.

Then we just need to tell the caldav lib to use our dummy storage:

```go
dummyStg := new(DummyStorage)
caldav.SetupStorage(dummyStg)
```

All the CRUD operations on resources will then be forwarded to our dummy storage.

The default storage used (if none is explicitly set) is the `data.FileStorage` which deals with resources as files in the File System.

The resources can be of two types: collection and non-collection. A collection resource is basically a resource that has children resources, but does not have any data content. A non-collection resource is a resource that does not have children, but has data. In the case of a file storage, collections correspond to directories and non-collection to plain files. The data of a caldav resource is all the info that shows up in the calendar client, in the [iCalendar](https://en.wikipedia.org/wiki/ICalendar) format.

### Configuration

You can set the caldav types your storage engine supports like so:

```go
caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VEVENT})
```

The default is `lib.VCALENDAR` and `lib.VEVENT`.

`caldav-go` will report these supported types to the client.

### Features

Please check the **CHANGELOG** to see specific features that are currently implemented.

### Contributing and testing

Everyone is welcome to contribute. Please raise an issue or pull request accordingly.

To run the tests:

```
./test.sh
```

### License

MIT License.

[RFC4791]: https://tools.ietf.org/html/rfc4791
