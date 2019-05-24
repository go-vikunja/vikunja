package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/global"
	"github.com/samedi/caldav-go/ixml"
)

type reportHandler struct {
	request  *http.Request
	response *Response
}

// See more at RFC4791#section-7.1
func (rh reportHandler) Handle() *Response {
	requestBody := readRequestBody(rh.request)
	header := headers{rh.request.Header}

	urlResource, found, err := global.Storage.GetShallowResource(rh.request.URL.Path)
	if !found {
		return rh.response.Set(http.StatusNotFound, "")
	} else if err != nil {
		return rh.response.SetError(err)
	}

	// read body string to xml struct
	var requestXML reportRootXML
	xml.Unmarshal([]byte(requestBody), &requestXML)

	// The resources to be reported are fetched by the type of the request. If it is
	// a `calendar-multiget`, the resources come based on a set of `hrefs` in the request body.
	// If it is a `calendar-query`, the resources are calculated based on set of filters in the request.
	var resourcesToReport []reportRes
	switch requestXML.XMLName {
	case ixml.CALENDAR_MULTIGET_TG:
		resourcesToReport, err = rh.fetchResourcesByList(urlResource, requestXML.Hrefs)
	case ixml.CALENDAR_QUERY_TG:
		resourcesToReport, err = rh.fetchResourcesByFilters(urlResource, requestXML.Filters)
	default:
		return rh.response.Set(http.StatusPreconditionFailed, "")
	}

	if err != nil {
		return rh.response.SetError(err)
	}

	multistatus := &multistatusResp{
		Minimal: header.IsMinimal(),
	}
	// for each href, build the multistatus responses
	for _, r := range resourcesToReport {
		propstats := multistatus.Propstats(r.resource, requestXML.Prop.Tags)
		multistatus.AddResponse(r.href, r.found, propstats)
	}

	if multistatus.Minimal {
		rh.response.SetHeader(HD_PREFERENCE_APPLIED, HD_PREFER_MINIMAL)
	}

	return rh.response.Set(207, multistatus.ToXML())
}

type reportPropXML struct {
	Tags []xml.Name `xml:",any"`
}

type reportRootXML struct {
	XMLName xml.Name
	Prop    reportPropXML   `xml:"DAV: prop"`
	Hrefs   []string        `xml:"DAV: href"`
	Filters reportFilterXML `xml:"urn:ietf:params:xml:ns:caldav filter"`
}

type reportFilterXML struct {
	XMLName      xml.Name
	InnerContent string `xml:",innerxml"`
}

func (rfXml reportFilterXML) toString() string {
	return fmt.Sprintf("<%s>%s</%s>", rfXml.XMLName.Local, rfXml.InnerContent, rfXml.XMLName.Local)
}

// Wraps a resource that has to be reported, either fetched by filters or by a list.
// Basically it contains the original requested `href`, the actual `resource` (can be nil)
// and if the `resource` was `found` or not
type reportRes struct {
	href     string
	resource *data.Resource
	found    bool
}

// The resources are fetched based on the origin resource and a set of filters.
// If the origin resource is a collection, the filters are checked against each of the collection's resources
// to see if they match. The collection's resources that match the filters are returned. The ones that will be returned
// are the resources that were not found (does not exist) and the ones that matched the filters. The ones that did not
// match the filter will not appear in the response result.
// If the origin resource is not a collection, the function just returns it and ignore any filter processing.
// [See RFC4791#section-7.8]
func (rh reportHandler) fetchResourcesByFilters(origin *data.Resource, filtersXML reportFilterXML) ([]reportRes, error) {
	// The list of resources that has to be reported back in the response.
	reps := []reportRes{}

	if origin.IsCollection() {
		filters, _ := data.ParseResourceFilters(filtersXML.toString())
		resources, err := global.Storage.GetResourcesByFilters(origin.Path, filters)

		if err != nil {
			return reps, err
		}

		for _, resource := range resources {
			reps = append(reps, reportRes{resource.Path, &resource, true})
		}
	} else {
		// the origin resource is not a collection, so returns just that as the result
		reps = append(reps, reportRes{origin.Path, origin, true})
	}

	return reps, nil
}

// The hrefs can come from (1) the request URL or (2) from the request body itself.
// If the origin resource from the URL points to a collection (2), we will check the request body
// to get the requested `hrefs` (resource paths). Each requested href has to be related to the collection.
// The ones that are not, we simply ignore them.
// If the resource from the URL is NOT a collection (1) we process the the report only for this resource
// and ignore any othre requested hrefs that might be present in the request body.
// [See RFC4791#section-7.9]
func (rh reportHandler) fetchResourcesByList(origin *data.Resource, requestedPaths []string) ([]reportRes, error) {
	reps := []reportRes{}

	if origin.IsCollection() {
		resources, err := global.Storage.GetResourcesByList(requestedPaths)

		if err != nil {
			return reps, err
		}

		// we put all the resources found in a map path -> resource.
		// this will be used later to query which requested resource was found
		// or not and mount the response
		resourcesMap := make(map[string]*data.Resource)
		for _, resource := range resources {
			r := resource
			resourcesMap[resource.Path] = &r
		}

		for _, requestedPath := range requestedPaths {
			// if the requested path does not belong to the origin collection, skip
			// ('belonging' means that the path's prefix is the same as the collection path)
			if !strings.HasPrefix(requestedPath, origin.Path) {
				continue
			}

			resource, found := resourcesMap[requestedPath]
			reps = append(reps, reportRes{requestedPath, resource, found})
		}
	} else {
		reps = append(reps, reportRes{origin.Path, origin, true})
	}

	return reps, nil
}
