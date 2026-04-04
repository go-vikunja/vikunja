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

package handler

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	AuthTypeUnknown int = iota
	AuthTypeUser
)

type McpHandler struct {
	EmptyStruct func() CObject
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Errorf("Error marshaling JSON: %v", err)
		return "{}"
	}
	return string(b)
}

func (c *McpHandler) getUser(request mcp.CallToolRequest) (*user.User, error) {

	token := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	jwtinf, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceSecret.GetString()), nil
	})
	if err != nil {
		return nil, err
	}

	claims := jwtinf.Claims.(jwt.MapClaims)
	typFloat, is := claims["type"].(float64)
	if !is {
		return nil, errors.New("invalid token")
	}
	typ := int(typFloat)

	if typ == AuthTypeUser {
		usr, e := user.GetUserFromClaims(claims)
		return usr, e
	}
	return nil, errors.New("invalid token")
}

func (c *McpHandler) getIndex() *reflect.StructField {
	t := reflect.TypeOf(c.EmptyStruct()).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if strings.Contains(field.Tag.Get("xorm"), "pk") {
			return &field
		}
	}
	return nil
}

// CObject is the definition of our object, holds the structs
type CObject interface {
	web.CRUDable
	web.Permissions
}

func (c *McpHandler) getTypeName() string {
	return strings.ToLower(reflect.TypeOf(c.EmptyStruct()).Elem().Name())
}

func (c *McpHandler) goToMCPType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Pointer, reflect.Struct, reflect.UnsafePointer:
		return "string"
	default:
		return "string"
	}
}

type List[T any] struct {
	Result T
	Cont   int
	Total  int64
}

func (c *McpHandler) getObjectMcpProperties() map[string]any {
	properties := map[string]any{}
	t := reflect.TypeOf(c.EmptyStruct()).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonKey := strings.Split(field.Tag.Get("json"), ",")[0]
		properties[jsonKey] = map[string]any{
			"type":        c.goToMCPType(field.Type),
			"description": field.Tag.Get("description")}
	}
	return properties
}
