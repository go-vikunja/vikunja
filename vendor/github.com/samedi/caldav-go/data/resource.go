package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/laurent22/ical-go"

	"github.com/samedi/caldav-go/files"
	"github.com/samedi/caldav-go/lib"
)

// ResourceInterface defines the main interface of a CalDAV resource object. This
// interface exists only to define the common resource operation and should not be custom-implemented.
// The default and canonical implementation is provided by `data.Resource`, convering all the commonalities.
// Any specifics in implementations should be handled by the `data.ResourceAdapter`.
type ResourceInterface interface {
	ComponentName() string
	StartTimeUTC() time.Time
	EndTimeUTC() time.Time
	Recurrences() []ResourceRecurrence
	HasProperty(propPath ...string) bool
	GetPropertyValue(propPath ...string) string
	HasPropertyParam(paramName ...string) bool
	GetPropertyParamValue(paramName ...string) string
}

// ResourceAdapter serves as the object to abstract all the specicities in different resources implementations.
// For example, the way to tell whether a resource is a collection or how to read its content differentiates
// on resources stored in the file system, coming from a relational DB or from the cloud as JSON. These differentiations
// should be covered by providing a specific implementation of the `ResourceAdapter` interface. So, depending on the current
// resource storage strategy, a matching resource adapter implementation should be provided whenever a new resource is initialized.
type ResourceAdapter interface {
	IsCollection() bool
	CalculateEtag() string
	GetContent() string
	GetContentSize() int64
	GetModTime() time.Time
}

// ResourceRecurrence represents a recurrence for a resource.
// NOTE: recurrences are not supported yet.
type ResourceRecurrence struct {
	StartTime time.Time
	EndTime   time.Time
}

// Resource represents the CalDAV resource. Basically, it has a name it's accessible based on path.
// A resource can be a collection, meaning it doesn't have any data content, but it has child resources.
// A non-collection is the actual resource which has the data in iCal format and which will feed the calendar.
// When visualizing the whole resources set in a tree representation, the collection resource would be the inner nodes and
// the non-collection would be the leaves.
type Resource struct {
	Name string
	Path string

	pathSplit []string
	adapter   ResourceAdapter

	emptyTime time.Time
}

// NewResource initializes a new `Resource` instance based on its path and the `ResourceAdapter` implementation to be used.
func NewResource(rawPath string, adp ResourceAdapter) Resource {
	pClean := lib.ToSlashPath(rawPath)
	pSplit := strings.Split(strings.Trim(pClean, "/"), "/")

	return Resource{
		Name:      pSplit[len(pSplit)-1],
		Path:      pClean,
		pathSplit: pSplit,
		adapter:   adp,
	}
}

// IsCollection tells whether a resource is a collection or not.
func (r *Resource) IsCollection() bool {
	return r.adapter.IsCollection()
}

// IsPrincipal tells whether a resource is the principal resource or not.
// A principal resource means it's a root resource.
func (r *Resource) IsPrincipal() bool {
	return len(r.pathSplit) <= 1
}

// ComponentName returns the type of the resource. VCALENDAR for collection resources, VEVENT otherwise.
func (r *Resource) ComponentName() string {
	if r.IsCollection() {
		return lib.VCALENDAR
	}

	return lib.VEVENT
}

// StartTimeUTC returns the start time in UTC of a VEVENT resource.
func (r *Resource) StartTimeUTC() time.Time {
	vevent := r.icalVEVENT()
	dtstart := vevent.PropDate(ical.DTSTART, r.emptyTime)

	if dtstart == r.emptyTime {
		log.Printf("WARNING: The property DTSTART was not found in the resource's ical data.\nResource path: %s", r.Path)
		return r.emptyTime
	}

	return dtstart.UTC()
}

// EndTimeUTC returns the end time in UTC of a VEVENT resource.
func (r *Resource) EndTimeUTC() time.Time {
	vevent := r.icalVEVENT()
	dtend := vevent.PropDate(ical.DTEND, r.emptyTime)

	// when the DTEND property is not present, we just add the DURATION (if any) to the DTSTART
	if dtend == r.emptyTime {
		duration := vevent.PropDuration(ical.DURATION)
		dtend = r.StartTimeUTC().Add(duration)
	}

	return dtend.UTC()
}

// Recurrences returns an array of resource recurrences.
// NOTE: Recurrences are not supported yet. An empty array will always be returned.
func (r *Resource) Recurrences() []ResourceRecurrence {
	// TODO: Implement. This server does not support iCal recurrences yet. We just return an empty array.
	return []ResourceRecurrence{}
}

// HasProperty tells whether the resource has the provided property in its iCal content.
// The path to the property should be provided in case of nested properties.
// Example, suppose the resource has this content:
//
// 	BEGIN:VCALENDAR
// 	BEGIN:VEVENT
// 	DTSTART:20160914T170000
// 	END:VEVENT
// 	END:VCALENDAR
//
// HasProperty("VEVENT", "DTSTART") => returns true
// HasProperty("VEVENT", "DTEND") => returns false
func (r *Resource) HasProperty(propPath ...string) bool {
	return r.GetPropertyValue(propPath...) != ""
}

// GetPropertyValue gets a property value from the resource's iCal content.
// The path to the property should be provided in case of nested properties.
// Example, suppose the resource has this content:
//
// 	BEGIN:VCALENDAR
// 	BEGIN:VEVENT
// 	DTSTART:20160914T170000
// 	END:VEVENT
// 	END:VCALENDAR
//
// GetPropertyValue("VEVENT", "DTSTART") => returns "20160914T170000"
// GetPropertyValue("VEVENT", "DTEND") => returns ""
func (r *Resource) GetPropertyValue(propPath ...string) string {
	if propPath[0] == ical.VCALENDAR {
		propPath = propPath[1:]
	}

	prop, _ := r.icalendar().DigProperty(propPath...)
	return prop
}

// HasPropertyParam tells whether the resource has the provided property param in its iCal content.
// The path to the param should be provided in case of nested params.
// Example, suppose the resource has this content:
//
// 	BEGIN:VCALENDAR
// 	BEGIN:VEVENT
// 	ATTENDEE;PARTSTAT=NEEDS-ACTION:FOO
// 	END:VEVENT
// 	END:VCALENDAR
//
// HasPropertyParam("VEVENT", "ATTENDEE", "PARTSTAT") => returns true
// HasPropertyParam("VEVENT", "ATTENDEE", "OTHER") => returns false
func (r *Resource) HasPropertyParam(paramPath ...string) bool {
	return r.GetPropertyParamValue(paramPath...) != ""
}

// GetPropertyParamValue gets a property param value from the resource's iCal content.
// The path to the param should be provided in case of nested params.
// Example, suppose the resource has this content:
//
// 	BEGIN:VCALENDAR
// 	BEGIN:VEVENT
// 	ATTENDEE;PARTSTAT=NEEDS-ACTION:FOO
// 	END:VEVENT
// 	END:VCALENDAR
//
// GetPropertyParamValue("VEVENT", "ATTENDEE", "PARTSTAT") => returns "NEEDS-ACTION"
// GetPropertyParamValue("VEVENT", "ATTENDEE", "OTHER") => returns ""
func (r *Resource) GetPropertyParamValue(paramPath ...string) string {
	if paramPath[0] == ical.VCALENDAR {
		paramPath = paramPath[1:]
	}

	param, _ := r.icalendar().DigParameter(paramPath...)
	return param
}

// GetEtag returns the ETag of the resource and a flag saying if the ETag is present.
// For collection resource, it returns an empty string and false.
func (r *Resource) GetEtag() (string, bool) {
	if r.IsCollection() {
		return "", false
	}

	return r.adapter.CalculateEtag(), true
}

// GetContentType returns the type of the content of the resource.
// Collection resources are "text/calendar". Non-collection resources are "text/calendar; component=vcalendar".
func (r *Resource) GetContentType() (string, bool) {
	if r.IsCollection() {
		return "text/calendar", true
	}

	return "text/calendar; component=vcalendar", true
}

// GetDisplayName returns the name/identifier of the resource.
func (r *Resource) GetDisplayName() (string, bool) {
	return r.Name, true
}

// GetContentData reads and returns the raw content of the resource as string and flag saying if the content was found.
// If the resource does not have content (like collection resource), it returns an empty string and false.
func (r *Resource) GetContentData() (string, bool) {
	data := r.adapter.GetContent()
	found := data != ""

	return data, found
}

// GetContentLength returns the length of the resource's content and flag saying if the length is present.
// If the resource does not have content (like collection resource), it returns an empty string and false.
func (r *Resource) GetContentLength() (string, bool) {
	// If its collection, it does not have any content, so mark it as not found
	if r.IsCollection() {
		return "", false
	}

	contentSize := r.adapter.GetContentSize()
	return strconv.FormatInt(contentSize, 10), true
}

// GetLastModified returns the last time the resource was modified. The returned time
// is returned formatted in the provided `format`.
func (r *Resource) GetLastModified(format string) (string, bool) {
	return r.adapter.GetModTime().Format(format), true
}

// GetOwner returns the owner of the resource. This is usually the principal resource associated (the root resource).
// If the resource does not have a owner (for example it's a principal resource alread), it returns an empty string.
func (r *Resource) GetOwner() (string, bool) {
	var owner string
	if len(r.pathSplit) > 1 {
		owner = r.pathSplit[0]
	} else {
		owner = ""
	}

	return owner, true
}

// GetOwnerPath returns the path to this resource's owner, or an empty string when the resource does not have any owner.
func (r *Resource) GetOwnerPath() (string, bool) {
	owner, _ := r.GetOwner()

	if owner != "" {
		return fmt.Sprintf("/%s/", owner), true
	}

	return "", false
}

// TODO: memoize
func (r *Resource) icalVEVENT() *ical.Node {
	vevent := r.icalendar().ChildByName(ical.VEVENT)

	// if nil, log it and return an empty vevent
	if vevent == nil {
		log.Printf("WARNING: The resource's ical data is missing the VEVENT property.\nResource path: %s", r.Path)

		return &ical.Node{
			Name: ical.VEVENT,
		}
	}

	return vevent
}

// TODO: memoize
func (r *Resource) icalendar() *ical.Node {
	data, found := r.GetContentData()

	if !found {
		log.Printf("WARNING: The resource's ical data does not have any data.\nResource path: %s", r.Path)
		return &ical.Node{
			Name: ical.VCALENDAR,
		}
	}

	icalNode, err := ical.ParseCalendar(data)
	if err != nil {
		log.Printf("ERROR: Could not parse the resource's ical data.\nError: %s.\nResource path: %s", err, r.Path)
		return &ical.Node{
			Name: ical.VCALENDAR,
		}
	}

	return icalNode
}

// FileResourceAdapter implements the `ResourceAdapter` for resources stored as files in the file system.
type FileResourceAdapter struct {
	finfo        os.FileInfo
	resourcePath string
}

// IsCollection tells whether the file resource is a directory or not.
func (adp *FileResourceAdapter) IsCollection() bool {
	return adp.finfo.IsDir()
}

// GetContent reads the file content and returns it as string. For collection resources (directories), it
// returns an empty string.
func (adp *FileResourceAdapter) GetContent() string {
	if adp.IsCollection() {
		return ""
	}

	data, err := ioutil.ReadFile(files.AbsPath(adp.resourcePath))
	if err != nil {
		log.Printf("ERROR: Could not read file content for the resource.\nError: %s.\nResource path: %s.", err, adp.resourcePath)
		return ""
	}

	return string(data)
}

// GetContentSize returns the content length.
func (adp *FileResourceAdapter) GetContentSize() int64 {
	return adp.finfo.Size()
}

// CalculateEtag calculates an ETag based on the file current modification status and returns it.
func (adp *FileResourceAdapter) CalculateEtag() string {
	// returns ETag as the concatenated hex values of a file's
	// modification time and size. This is not a reliable synchronization
	// mechanism for directories, so for collections we return empty.
	if adp.IsCollection() {
		return ""
	}

	fi := adp.finfo
	return fmt.Sprintf(`"%x%x"`, fi.ModTime().UnixNano(), fi.Size())
}

// GetModTime returns the time when the file was last modified.
func (adp *FileResourceAdapter) GetModTime() time.Time {
	return adp.finfo.ModTime()
}
