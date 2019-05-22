package data

import (
	"errors"
	"github.com/beevik/etree"
	"log"
	"strings"
	"time"

	"github.com/samedi/caldav-go/lib"
)

const (
	TAG_FILTER         = "filter"
	TAG_COMP_FILTER    = "comp-filter"
	TAG_PROP_FILTER    = "prop-filter"
	TAG_PARAM_FILTER   = "param-filter"
	TAG_TIME_RANGE     = "time-range"
	TAG_TEXT_MATCH     = "text-match"
	TAG_IS_NOT_DEFINED = "is-not-defined"

	// From the RFC, the time range `start` and `end` attributes MUST be in UTC and in this specific format
	FILTER_TIME_FORMAT = "20060102T150405Z"
)

// ResourceFilter represents filters to filter out resources.
// Filters are basically a set of rules used to retrieve a range of resources.
// It is used primarily on REPORT requests and is described in details in RFC4791#7.8.
type ResourceFilter struct {
	name      string
	text      string
	attrs     map[string]string
	children  []ResourceFilter // collection of child filters.
	etreeElem *etree.Element   // holds the parsed XML node/tag as an `etree` element.
}

// ParseResourceFilters initializes a new `ResourceFilter` object from a snippet of XML string.
func ParseResourceFilters(xml string) (*ResourceFilter, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		log.Printf("ERROR: Could not parse filter from XML string. XML:\n%s", xml)
		return new(ResourceFilter), err
	}

	// Right now we're searching for a <filter> tag to initialize the filter struct from it.
	// It SHOULD be a valid XML CALDAV:filter tag (RFC4791#9.7). We're not checking namespaces yet.
	// TODO: check for XML namespaces and restrict it to accept only CALDAV:filter tag.
	elem := doc.FindElement("//" + TAG_FILTER)
	if elem == nil {
		log.Printf("WARNING: The filter XML should contain a <%s> element. XML:\n%s", TAG_FILTER, xml)
		return new(ResourceFilter), errors.New("invalid XML filter")
	}

	filter := newFilterFromEtreeElem(elem)
	return &filter, nil
}

func newFilterFromEtreeElem(elem *etree.Element) ResourceFilter {
	// init filter from etree element
	filter := ResourceFilter{
		name:      elem.Tag,
		text:      strings.TrimSpace(elem.Text()),
		etreeElem: elem,
		attrs:     make(map[string]string),
	}

	// set attributes
	for _, attr := range elem.Attr {
		filter.attrs[attr.Key] = attr.Value
	}

	return filter
}

// Attr searches an attribute by its name in the list of filter attributes and returns it.
func (f *ResourceFilter) Attr(attrName string) string {
	return f.attrs[attrName]
}

// TimeAttr searches and returns a filter attribute as a `time.Time` object.
func (f *ResourceFilter) TimeAttr(attrName string) *time.Time {

	t, err := time.Parse(FILTER_TIME_FORMAT, f.attrs[attrName])
	if err != nil {
		return nil
	}

	return &t
}

// GetTimeRangeFilter checks if the current filter has a child "time-range" filter and
// returns it (wrapped in a `ResourceFilter` type). It returns nil if the current filter does
// not contain any "time-range" filter.
func (f *ResourceFilter) GetTimeRangeFilter() *ResourceFilter {
	return f.findChild(TAG_TIME_RANGE, true)
}

// Match returns whether a provided resource matches the filters.
func (f *ResourceFilter) Match(target ResourceInterface) bool {
	if f.name == TAG_FILTER {
		return f.rootFilterMatch(target)
	}

	return false
}

func (f *ResourceFilter) rootFilterMatch(target ResourceInterface) bool {
	if f.isEmpty() {
		return false
	}

	return f.rootChildrenMatch(target)
}

// checks if all the root's child filters match the target resource
func (f *ResourceFilter) rootChildrenMatch(target ResourceInterface) bool {
	scope := []string{}

	for _, child := range f.getChildren() {
		// root filters only accept comp filters as children
		if child.name != TAG_COMP_FILTER || !child.compMatch(target, scope) {
			return false
		}
	}

	return true
}

// See RFC4791-9.7.1.
func (f *ResourceFilter) compMatch(target ResourceInterface, scope []string) bool {
	targetComp := target.ComponentName()
	compName := f.attrs["name"]

	if f.isEmpty() {
		// Point #1 of RFC4791#9.7.1
		return compName == targetComp
	} else if f.contains(TAG_IS_NOT_DEFINED) {
		// Point #2 of RFC4791#9.7.1
		return compName != targetComp
	} else {
		// check each child of the current filter if they all match.
		childrenScope := append(scope, compName)
		return f.compChildrenMatch(target, childrenScope)
	}
}

// checks if all the comp's child filters match the target resource
func (f *ResourceFilter) compChildrenMatch(target ResourceInterface, scope []string) bool {
	for _, child := range f.getChildren() {
		var match bool

		switch child.name {
		case TAG_TIME_RANGE:
			// Point #3 of RFC4791#9.7.1
			match = child.timeRangeMatch(target)
		case TAG_PROP_FILTER:
			// Point #4 of RFC4791#9.7.1
			match = child.propMatch(target, scope)
		case TAG_COMP_FILTER:
			// Point #4 of RFC4791#9.7.1
			match = child.compMatch(target, scope)
		}

		if !match {
			return false
		}
	}

	return true
}

// See RFC4791-9.9
func (f *ResourceFilter) timeRangeMatch(target ResourceInterface) bool {
	startAttr := f.attrs["start"]
	endAttr := f.attrs["end"]

	// at least one of the two MUST be present
	if startAttr == "" && endAttr == "" {
		// if both of them are missing, return false
		return false
	} else if startAttr == "" {
		// if missing only the `start`, set it open ended to the left
		startAttr = "00010101T000000Z"
	} else if endAttr == "" {
		// if missing only the `end`, set it open ended to the right
		endAttr = "99991231T235959Z"
	}

	// The logic below is only applicable for VEVENT components. So
	// we return false if the resource is not a VEVENT component.
	if target.ComponentName() != lib.VEVENT {
		return false
	}

	rangeStart, err := time.Parse(FILTER_TIME_FORMAT, startAttr)
	if err != nil {
		log.Printf("ERROR: Could not parse start time in time-range filter.\nError: %s.\nStart attr: %s", err, startAttr)
		return false
	}

	rangeEnd, err := time.Parse(FILTER_TIME_FORMAT, endAttr)
	if err != nil {
		log.Printf("ERROR: Could not parse end time in time-range filter.\nError: %s.\nEnd attr: %s", err, endAttr)
		return false
	}

	// the following logic is inferred from the rules table for VEVENT components,
	// described in RFC4791-9.9.
	overlapRange := func(dtStart, dtEnd, rangeStart, rangeEnd time.Time) bool {
		if dtStart.Equal(dtEnd) {
			// Lines 3 and 4 of the table deal when the DTSTART and DTEND dates are equals.
			// In this case we use the rule: (start <= DTSTART && end > DTSTART)
			return (rangeStart.Before(dtStart) || rangeStart.Equal(dtStart)) && rangeEnd.After(dtStart)
		} else {
			// Lines 1, 2 and 6 of the table deal when the DTSTART and DTEND dates are different.
			// In this case we use the rule: (start < DTEND && end > DTSTART)
			return rangeStart.Before(dtEnd) && rangeEnd.After(dtStart)
		}
	}

	// first we check each of the target recurrences (if any).
	for _, recurrence := range target.Recurrences() {
		// if any of them overlap the filter range, we return true right away
		if overlapRange(recurrence.StartTime, recurrence.EndTime, rangeStart, rangeEnd) {
			return true
		}
	}

	// if none of the recurrences match, we just return if the actual
	// resource's `start` and `end` times match the filter range
	return overlapRange(target.StartTimeUTC(), target.EndTimeUTC(), rangeStart, rangeEnd)
}

// See RFC4791-9.7.2.
func (f *ResourceFilter) propMatch(target ResourceInterface, scope []string) bool {
	propName := f.attrs["name"]
	propPath := append(scope, propName)

	if f.isEmpty() {
		// Point #1 of RFC4791#9.7.2
		return target.HasProperty(propPath...)
	} else if f.contains(TAG_IS_NOT_DEFINED) {
		// Point #2 of RFC4791#9.7.2
		return !target.HasProperty(propPath...)
	} else {
		// check each child of the current filter if they all match.
		return f.propChildrenMatch(target, propPath)
	}
}

// checks if all the prop's child filters match the target resource
func (f *ResourceFilter) propChildrenMatch(target ResourceInterface, propPath []string) bool {
	for _, child := range f.getChildren() {
		var match bool

		switch child.name {
		case TAG_TIME_RANGE:
			// Point #3 of RFC4791#9.7.2
			// TODO: this point is not very clear on how to match time range against properties.
			// So we're returning `false` in the meantime.
			match = false
		case TAG_TEXT_MATCH:
			// Point #4 of RFC4791#9.7.2
			propText := target.GetPropertyValue(propPath...)
			match = child.textMatch(propText)
		case TAG_PARAM_FILTER:
			// Point #4 of RFC4791#9.7.2
			match = child.paramMatch(target, propPath)
		}

		if !match {
			return false
		}
	}

	return true
}

// See RFC4791-9.7.3
func (f *ResourceFilter) paramMatch(target ResourceInterface, parentPropPath []string) bool {
	paramName := f.attrs["name"]
	paramPath := append(parentPropPath, paramName)

	if f.isEmpty() {
		// Point #1 of RFC4791#9.7.3
		return target.HasPropertyParam(paramPath...)
	} else if f.contains(TAG_IS_NOT_DEFINED) {
		// Point #2 of RFC4791#9.7.3
		return !target.HasPropertyParam(paramPath...)
	} else {
		child := f.getChildren()[0]
		// param filters can also have (only-one) nested text-match filter
		if child.name == TAG_TEXT_MATCH {
			paramValue := target.GetPropertyParamValue(paramPath...)
			return child.textMatch(paramValue)
		}
	}

	return false
}

// See RFC4791-9.7.5
func (f *ResourceFilter) textMatch(targetText string) bool {
	// TODO: collations are not being considered/supported yet.
	// Texts are lowered to be case-insensitive, almost as the "i;ascii-casemap" value.

	targetText = strings.ToLower(targetText)
	expectedSubstr := strings.ToLower(f.text)

	match := strings.Contains(targetText, expectedSubstr)

	if f.attrs["negate-condition"] == "yes" {
		return !match
	}

	return match
}

func (f *ResourceFilter) isEmpty() bool {
	return len(f.getChildren()) == 0 && f.text == ""
}

func (f *ResourceFilter) contains(filterName string) bool {
	if f.findChild(filterName, false) != nil {
		return true
	}

	return false
}

func (f *ResourceFilter) findChild(filterName string, dig bool) *ResourceFilter {
	for _, child := range f.getChildren() {
		if child.name == filterName {
			return &child
		}

		if !dig {
			continue
		}

		dugChild := child.findChild(filterName, true)

		if dugChild != nil {
			return dugChild
		}
	}

	return nil
}

// lazy evaluation of the child filters
func (f *ResourceFilter) getChildren() []ResourceFilter {
	if f.children == nil {
		f.children = []ResourceFilter{}

		for _, childElem := range f.etreeElem.ChildElements() {
			childFilter := newFilterFromEtreeElem(childElem)
			f.children = append(f.children, childFilter)
		}
	}

	return f.children
}
