package handlers

import (
	"encoding/xml"
	"fmt"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/global"
	"github.com/samedi/caldav-go/ixml"
	"github.com/samedi/caldav-go/lib"
	"net/http"
)

// Wraps a multistatus response. It contains the set of `Responses`
// that will serve to build the final XML. Multistatus responses are
// used by the REPORT and PROPFIND methods.
type multistatusResp struct {
	// The set of multistatus responses used to build each of the <DAV:response> nodes.
	Responses []msResponse
	// Flag that XML should be minimal or not
	// [defined in the draft https://tools.ietf.org/html/draft-murchison-webdav-prefer-05]
	Minimal bool
}

type msResponse struct {
	Href      string
	Found     bool
	Propstats msPropstats
}

type msPropstats map[int]msProps

// Adds a msProp to the map with the key being the prop status.
func (stats msPropstats) Add(prop msProp) {
	stats[prop.Status] = append(stats[prop.Status], prop)
}

func (stats msPropstats) Clone() msPropstats {
	clone := make(msPropstats)

	for k, v := range stats {
		clone[k] = v
	}

	return clone
}

type msProps []msProp

type msProp struct {
	Tag      xml.Name
	Content  string
	Contents []string
	Status   int
}

// Function that processes all the required props for a given resource.
// ## Params
// resource: the target calendar resource.
// reqprops: set of required props that must be processed for the resource.
// ## Returns
// The set of props (msProp) processed. Each prop is mapped to a HTTP status code.
// So if a prop is found and processed ok, it'll be mapped to 200. If it's not found,
// it'll be mapped to 404, and so on.
func (ms *multistatusResp) Propstats(resource *data.Resource, reqprops []xml.Name) msPropstats {
	if resource == nil {
		return nil
	}

	result := make(msPropstats)

	for _, ptag := range reqprops {
		pvalue := msProp{
			Tag:    ptag,
			Status: http.StatusOK,
		}

		pfound := false
		switch ptag {
		case ixml.CALENDAR_DATA_TG:
			pvalue.Content, pfound = resource.GetContentData()
			if pfound {
				pvalue.Content = ixml.EscapeText(pvalue.Content)
			}
		case ixml.GET_ETAG_TG:
			pvalue.Content, pfound = resource.GetEtag()
		case ixml.GET_CONTENT_TYPE_TG:
			pvalue.Content, pfound = resource.GetContentType()
		case ixml.GET_CONTENT_LENGTH_TG:
			pvalue.Content, pfound = resource.GetContentLength()
		case ixml.DISPLAY_NAME_TG:
			pvalue.Content, pfound = resource.GetDisplayName()
			if pfound {
				pvalue.Content = ixml.EscapeText(pvalue.Content)
			}
		case ixml.GET_LAST_MODIFIED_TG:
			pvalue.Content, pfound = resource.GetLastModified(http.TimeFormat)
		case ixml.OWNER_TG:
			pvalue.Content, pfound = resource.GetOwnerPath()
		case ixml.GET_CTAG_TG:
			pvalue.Content, pfound = resource.GetEtag()
		case ixml.PRINCIPAL_URL_TG,
			ixml.PRINCIPAL_COLLECTION_SET_TG,
			ixml.CALENDAR_USER_ADDRESS_SET_TG,
			ixml.CALENDAR_HOME_SET_TG:
			pvalue.Content, pfound = ixml.HrefTag(resource.Path), true
		case ixml.RESOURCE_TYPE_TG:
			if resource.IsCollection() {
				pvalue.Content, pfound = ixml.Tag(ixml.COLLECTION_TG, "")+ixml.Tag(ixml.CALENDAR_TG, ""), true

				if resource.IsPrincipal() {
					pvalue.Content += ixml.Tag(ixml.PRINCIPAL_TG, "")
				}
			} else {
				// resourcetype must be returned empty for non-collection elements
				pvalue.Content, pfound = "", true
			}
		case ixml.CURRENT_USER_PRINCIPAL_TG:
			if global.User != nil {
				path := fmt.Sprintf("/%s/", global.User.Name)
				pvalue.Content, pfound = ixml.HrefTag(path), true
			}
		case ixml.SUPPORTED_CALENDAR_COMPONENT_SET_TG:
			if resource.IsCollection() {
				for _, component := range global.SupportedComponents {
					// TODO: use ixml somehow to build the below tag
					compTag := fmt.Sprintf(`<C:comp name="%s"/>`, component)
					pvalue.Contents = append(pvalue.Contents, compTag)
				}
				pfound = true
			}
		}

		if !pfound {
			pvalue.Status = http.StatusNotFound
		}

		result.Add(pvalue)
	}

	return result
}

// Adds a new `msResponse` to the `Responses` array.
func (ms *multistatusResp) AddResponse(href string, found bool, propstats msPropstats) {
	ms.Responses = append(ms.Responses, msResponse{
		Href:      href,
		Found:     found,
		Propstats: propstats,
	})
}

func (ms *multistatusResp) ToXML() string {
	// init multistatus
	var bf lib.StringBuffer
	bf.Write(`<?xml version="1.0" encoding="UTF-8"?>`)
	bf.Write(`<D:multistatus %s>`, ixml.Namespaces())

	// iterate over event hrefs and build multistatus XML on the fly
	for _, response := range ms.Responses {
		bf.Write("<D:response>")
		bf.Write(ixml.HrefTag(response.Href))

		if response.Found {
			propstats := response.Propstats.Clone()

			if ms.Minimal {
				delete(propstats, http.StatusNotFound)

				if len(propstats) == 0 {
					bf.Write("<D:propstat>")
					bf.Write("<D:prop/>")
					bf.Write(ixml.StatusTag(http.StatusOK))
					bf.Write("</D:propstat>")
					bf.Write("</D:response>")

					continue
				}
			}

			for status, props := range propstats {
				bf.Write("<D:propstat>")
				bf.Write("<D:prop>")
				for _, prop := range props {
					bf.Write(ms.propToXML(prop))
				}
				bf.Write("</D:prop>")
				bf.Write(ixml.StatusTag(status))
				bf.Write("</D:propstat>")
			}
		} else {
			// if does not find the resource set 404
			bf.Write(ixml.StatusTag(http.StatusNotFound))
		}
		bf.Write("</D:response>")
	}
	bf.Write("</D:multistatus>")

	return bf.String()
}

func (ms *multistatusResp) propToXML(prop msProp) string {
	for _, content := range prop.Contents {
		prop.Content += content
	}
	xmlString := ixml.Tag(prop.Tag, prop.Content)
	return xmlString
}
