// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package apiv2

import (
	"reflect"

	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

const clearOutsideWebhooksDoc = "What to do with webhooks created by users who lose access to the project or its child projects: true deletes them, false keeps them. Their task assignments and subscriptions are removed either way. When omitted and any such webhook exists, the request fails with 409 and code 3018 without changing anything."

// clearOutsideWebhooksParam tells an omitted query value from false; Huma rejects pointer params.
type clearOutsideWebhooksParam struct {
	value bool
	set   bool
}

func (p *clearOutsideWebhooksParam) Receiver() reflect.Value {
	return reflect.ValueOf(&p.value).Elem()
}

func (p *clearOutsideWebhooksParam) OnParamSet(isSet bool, _ any) {
	p.set = isSet
}

func (*clearOutsideWebhooksParam) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{Type: huma.TypeBoolean, Description: clearOutsideWebhooksDoc}
}

func (p *clearOutsideWebhooksParam) option() models.OutsideWebhooks {
	if !p.set {
		return models.OutsideWebhooksFromChoice(nil)
	}
	return models.OutsideWebhooksFromChoice(&p.value)
}

func outsideWebhooksResponses(api huma.API) map[string]*huma.Response {
	return map[string]*huma.Response{
		"default": defaultErrorResponse(api),
		"409": {
			Description: "clear_outside_webhooks was omitted and the change would delete webhooks of users who lose access (code 3018). i18n_params.count holds how many. Nothing was changed; repeat the request with clear_outside_webhooks set.",
			Content:     defaultErrorResponse(api).Content,
		},
	}
}
