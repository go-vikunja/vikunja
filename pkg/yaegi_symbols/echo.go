




package yaegi_symbols

import (
	"go/constant"
	"go/token"
	"io"
	"github.com/labstack/echo/v5"
	"reflect"
)

func init() {
	Symbols["github.com/labstack/echo/v5/echo"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"BindBody": reflect.ValueOf(echo.BindBody),
			"BindHeaders": reflect.ValueOf(echo.BindHeaders),
			"BindPathValues": reflect.ValueOf(echo.BindPathValues),
			"BindQueryParams": reflect.ValueOf(echo.BindQueryParams),
			"ContextKeyHeaderAllow": reflect.ValueOf(constant.MakeFromLiteral("\"echo_header_allow\"", token.STRING, 0)),
			"DefaultHTTPErrorHandler": reflect.ValueOf(echo.DefaultHTTPErrorHandler),
			"ErrBadGateway": reflect.ValueOf(&echo.ErrBadGateway).Elem(),
			"ErrBadRequest": reflect.ValueOf(&echo.ErrBadRequest).Elem(),
			"ErrCookieNotFound": reflect.ValueOf(&echo.ErrCookieNotFound).Elem(),
			"ErrForbidden": reflect.ValueOf(&echo.ErrForbidden).Elem(),
			"ErrInternalServerError": reflect.ValueOf(&echo.ErrInternalServerError).Elem(),
			"ErrInvalidCertOrKeyType": reflect.ValueOf(&echo.ErrInvalidCertOrKeyType).Elem(),
			"ErrInvalidKeyType": reflect.ValueOf(&echo.ErrInvalidKeyType).Elem(),
			"ErrInvalidListenerNetwork": reflect.ValueOf(&echo.ErrInvalidListenerNetwork).Elem(),
			"ErrInvalidRedirectCode": reflect.ValueOf(&echo.ErrInvalidRedirectCode).Elem(),
			"ErrMethodNotAllowed": reflect.ValueOf(&echo.ErrMethodNotAllowed).Elem(),
			"ErrNonExistentKey": reflect.ValueOf(&echo.ErrNonExistentKey).Elem(),
			"ErrNotFound": reflect.ValueOf(&echo.ErrNotFound).Elem(),
			"ErrRendererNotRegistered": reflect.ValueOf(&echo.ErrRendererNotRegistered).Elem(),
			"ErrRequestTimeout": reflect.ValueOf(&echo.ErrRequestTimeout).Elem(),
			"ErrServiceUnavailable": reflect.ValueOf(&echo.ErrServiceUnavailable).Elem(),
			"ErrStatusRequestEntityTooLarge": reflect.ValueOf(&echo.ErrStatusRequestEntityTooLarge).Elem(),
			"ErrTooManyRequests": reflect.ValueOf(&echo.ErrTooManyRequests).Elem(),
			"ErrUnauthorized": reflect.ValueOf(&echo.ErrUnauthorized).Elem(),
			"ErrUnsupportedMediaType": reflect.ValueOf(&echo.ErrUnsupportedMediaType).Elem(),
			"ErrValidatorNotRegistered": reflect.ValueOf(&echo.ErrValidatorNotRegistered).Elem(),
			"ExtractIPDirect": reflect.ValueOf(echo.ExtractIPDirect),
			"ExtractIPFromRealIPHeader": reflect.ValueOf(echo.ExtractIPFromRealIPHeader),
			"ExtractIPFromXFFHeader": reflect.ValueOf(echo.ExtractIPFromXFFHeader),
			"FormFieldBinder": reflect.ValueOf(echo.FormFieldBinder),
			"HandlerName": reflect.ValueOf(echo.HandlerName),
			"HeaderAccept": reflect.ValueOf(constant.MakeFromLiteral("\"Accept\"", token.STRING, 0)),
			"HeaderAcceptEncoding": reflect.ValueOf(constant.MakeFromLiteral("\"Accept-Encoding\"", token.STRING, 0)),
			"HeaderAccessControlAllowCredentials": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Allow-Credentials\"", token.STRING, 0)),
			"HeaderAccessControlAllowHeaders": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Allow-Headers\"", token.STRING, 0)),
			"HeaderAccessControlAllowMethods": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Allow-Methods\"", token.STRING, 0)),
			"HeaderAccessControlAllowOrigin": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Allow-Origin\"", token.STRING, 0)),
			"HeaderAccessControlExposeHeaders": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Expose-Headers\"", token.STRING, 0)),
			"HeaderAccessControlMaxAge": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Max-Age\"", token.STRING, 0)),
			"HeaderAccessControlRequestHeaders": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Request-Headers\"", token.STRING, 0)),
			"HeaderAccessControlRequestMethod": reflect.ValueOf(constant.MakeFromLiteral("\"Access-Control-Request-Method\"", token.STRING, 0)),
			"HeaderAllow": reflect.ValueOf(constant.MakeFromLiteral("\"Allow\"", token.STRING, 0)),
			"HeaderAuthorization": reflect.ValueOf(constant.MakeFromLiteral("\"Authorization\"", token.STRING, 0)),
			"HeaderCacheControl": reflect.ValueOf(constant.MakeFromLiteral("\"Cache-Control\"", token.STRING, 0)),
			"HeaderConnection": reflect.ValueOf(constant.MakeFromLiteral("\"Connection\"", token.STRING, 0)),
			"HeaderContentDisposition": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Disposition\"", token.STRING, 0)),
			"HeaderContentEncoding": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Encoding\"", token.STRING, 0)),
			"HeaderContentLength": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Length\"", token.STRING, 0)),
			"HeaderContentSecurityPolicy": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Security-Policy\"", token.STRING, 0)),
			"HeaderContentSecurityPolicyReportOnly": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Security-Policy-Report-Only\"", token.STRING, 0)),
			"HeaderContentType": reflect.ValueOf(constant.MakeFromLiteral("\"Content-Type\"", token.STRING, 0)),
			"HeaderCookie": reflect.ValueOf(constant.MakeFromLiteral("\"Cookie\"", token.STRING, 0)),
			"HeaderIfModifiedSince": reflect.ValueOf(constant.MakeFromLiteral("\"If-Modified-Since\"", token.STRING, 0)),
			"HeaderLastModified": reflect.ValueOf(constant.MakeFromLiteral("\"Last-Modified\"", token.STRING, 0)),
			"HeaderLocation": reflect.ValueOf(constant.MakeFromLiteral("\"Location\"", token.STRING, 0)),
			"HeaderOrigin": reflect.ValueOf(constant.MakeFromLiteral("\"Origin\"", token.STRING, 0)),
			"HeaderReferrerPolicy": reflect.ValueOf(constant.MakeFromLiteral("\"Referrer-Policy\"", token.STRING, 0)),
			"HeaderRetryAfter": reflect.ValueOf(constant.MakeFromLiteral("\"Retry-After\"", token.STRING, 0)),
			"HeaderSecFetchSite": reflect.ValueOf(constant.MakeFromLiteral("\"Sec-Fetch-Site\"", token.STRING, 0)),
			"HeaderServer": reflect.ValueOf(constant.MakeFromLiteral("\"Server\"", token.STRING, 0)),
			"HeaderSetCookie": reflect.ValueOf(constant.MakeFromLiteral("\"Set-Cookie\"", token.STRING, 0)),
			"HeaderStrictTransportSecurity": reflect.ValueOf(constant.MakeFromLiteral("\"Strict-Transport-Security\"", token.STRING, 0)),
			"HeaderUpgrade": reflect.ValueOf(constant.MakeFromLiteral("\"Upgrade\"", token.STRING, 0)),
			"HeaderVary": reflect.ValueOf(constant.MakeFromLiteral("\"Vary\"", token.STRING, 0)),
			"HeaderWWWAuthenticate": reflect.ValueOf(constant.MakeFromLiteral("\"WWW-Authenticate\"", token.STRING, 0)),
			"HeaderXCSRFToken": reflect.ValueOf(constant.MakeFromLiteral("\"X-CSRF-Token\"", token.STRING, 0)),
			"HeaderXContentTypeOptions": reflect.ValueOf(constant.MakeFromLiteral("\"X-Content-Type-Options\"", token.STRING, 0)),
			"HeaderXCorrelationID": reflect.ValueOf(constant.MakeFromLiteral("\"X-Correlation-Id\"", token.STRING, 0)),
			"HeaderXForwardedFor": reflect.ValueOf(constant.MakeFromLiteral("\"X-Forwarded-For\"", token.STRING, 0)),
			"HeaderXForwardedProto": reflect.ValueOf(constant.MakeFromLiteral("\"X-Forwarded-Proto\"", token.STRING, 0)),
			"HeaderXForwardedProtocol": reflect.ValueOf(constant.MakeFromLiteral("\"X-Forwarded-Protocol\"", token.STRING, 0)),
			"HeaderXForwardedSsl": reflect.ValueOf(constant.MakeFromLiteral("\"X-Forwarded-Ssl\"", token.STRING, 0)),
			"HeaderXFrameOptions": reflect.ValueOf(constant.MakeFromLiteral("\"X-Frame-Options\"", token.STRING, 0)),
			"HeaderXHTTPMethodOverride": reflect.ValueOf(constant.MakeFromLiteral("\"X-HTTP-Method-Override\"", token.STRING, 0)),
			"HeaderXRealIP": reflect.ValueOf(constant.MakeFromLiteral("\"X-Real-Ip\"", token.STRING, 0)),
			"HeaderXRequestID": reflect.ValueOf(constant.MakeFromLiteral("\"X-Request-Id\"", token.STRING, 0)),
			"HeaderXRequestedWith": reflect.ValueOf(constant.MakeFromLiteral("\"X-Requested-With\"", token.STRING, 0)),
			"HeaderXUrlScheme": reflect.ValueOf(constant.MakeFromLiteral("\"X-Url-Scheme\"", token.STRING, 0)),
			"HeaderXXSSProtection": reflect.ValueOf(constant.MakeFromLiteral("\"X-XSS-Protection\"", token.STRING, 0)),
			"MIMEApplicationForm": reflect.ValueOf(constant.MakeFromLiteral("\"application/x-www-form-urlencoded\"", token.STRING, 0)),
			"MIMEApplicationJSON": reflect.ValueOf(constant.MakeFromLiteral("\"application/json\"", token.STRING, 0)),
			"MIMEApplicationJSONCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"application/json; charset=UTF-8\"", token.STRING, 0)),
			"MIMEApplicationJavaScript": reflect.ValueOf(constant.MakeFromLiteral("\"application/javascript\"", token.STRING, 0)),
			"MIMEApplicationJavaScriptCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"application/javascript; charset=UTF-8\"", token.STRING, 0)),
			"MIMEApplicationMsgpack": reflect.ValueOf(constant.MakeFromLiteral("\"application/msgpack\"", token.STRING, 0)),
			"MIMEApplicationProtobuf": reflect.ValueOf(constant.MakeFromLiteral("\"application/protobuf\"", token.STRING, 0)),
			"MIMEApplicationXML": reflect.ValueOf(constant.MakeFromLiteral("\"application/xml\"", token.STRING, 0)),
			"MIMEApplicationXMLCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"application/xml; charset=UTF-8\"", token.STRING, 0)),
			"MIMEMultipartForm": reflect.ValueOf(constant.MakeFromLiteral("\"multipart/form-data\"", token.STRING, 0)),
			"MIMEOctetStream": reflect.ValueOf(constant.MakeFromLiteral("\"application/octet-stream\"", token.STRING, 0)),
			"MIMETextHTML": reflect.ValueOf(constant.MakeFromLiteral("\"text/html\"", token.STRING, 0)),
			"MIMETextHTMLCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"text/html; charset=UTF-8\"", token.STRING, 0)),
			"MIMETextPlain": reflect.ValueOf(constant.MakeFromLiteral("\"text/plain\"", token.STRING, 0)),
			"MIMETextPlainCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"text/plain; charset=UTF-8\"", token.STRING, 0)),
			"MIMETextXML": reflect.ValueOf(constant.MakeFromLiteral("\"text/xml\"", token.STRING, 0)),
			"MIMETextXMLCharsetUTF8": reflect.ValueOf(constant.MakeFromLiteral("\"text/xml; charset=UTF-8\"", token.STRING, 0)),
			"MethodNotAllowedRouteName": reflect.ValueOf(constant.MakeFromLiteral("\"echo_route_method_not_allowed_name\"", token.STRING, 0)),
			"MustSubFS": reflect.ValueOf(echo.MustSubFS),
			"New": reflect.ValueOf(echo.New),
			"NewBindingError": reflect.ValueOf(echo.NewBindingError),
			"NewConcurrentRouter": reflect.ValueOf(echo.NewConcurrentRouter),
			"NewContext": reflect.ValueOf(echo.NewContext),
			"NewHTTPError": reflect.ValueOf(echo.NewHTTPError),
			"NewResponse": reflect.ValueOf(echo.NewResponse),
			"NewRouter": reflect.ValueOf(echo.NewRouter),
			"NewVirtualHostHandler": reflect.ValueOf(echo.NewVirtualHostHandler),
			"NewWithConfig": reflect.ValueOf(echo.NewWithConfig),
			"NotFoundRouteName": reflect.ValueOf(constant.MakeFromLiteral("\"echo_route_not_found_name\"", token.STRING, 0)),
			"PROPFIND": reflect.ValueOf(constant.MakeFromLiteral("\"PROPFIND\"", token.STRING, 0)),
			"PathValuesBinder": reflect.ValueOf(echo.PathValuesBinder),
			"QueryParamsBinder": reflect.ValueOf(echo.QueryParamsBinder),
			"REPORT": reflect.ValueOf(constant.MakeFromLiteral("\"REPORT\"", token.STRING, 0)),
			"RouteAny": reflect.ValueOf(constant.MakeFromLiteral("\"echo_route_any\"", token.STRING, 0)),
			"RouteNotFound": reflect.ValueOf(constant.MakeFromLiteral("\"echo_route_not_found\"", token.STRING, 0)),
			"StaticDirectoryHandler": reflect.ValueOf(echo.StaticDirectoryHandler),
			"StaticFileHandler": reflect.ValueOf(echo.StaticFileHandler),
			"TimeLayoutUnixTime": reflect.ValueOf(echo.TimeLayoutUnixTime),
			"TimeLayoutUnixTimeMilli": reflect.ValueOf(echo.TimeLayoutUnixTimeMilli),
			"TimeLayoutUnixTimeNano": reflect.ValueOf(echo.TimeLayoutUnixTimeNano),
			"TrustIPRange": reflect.ValueOf(echo.TrustIPRange),
			"TrustLinkLocal": reflect.ValueOf(echo.TrustLinkLocal),
			"TrustLoopback": reflect.ValueOf(echo.TrustLoopback),
			"TrustPrivateNet": reflect.ValueOf(echo.TrustPrivateNet),
			"UnwrapResponse": reflect.ValueOf(echo.UnwrapResponse),
			"Version": reflect.ValueOf(constant.MakeFromLiteral("\"5.0.0\"", token.STRING, 0)),
			"WrapHandler": reflect.ValueOf(echo.WrapHandler),
			"WrapMiddleware": reflect.ValueOf(echo.WrapMiddleware),
			
		// type definitions
		"AddRouteError": reflect.ValueOf((*echo.AddRouteError)(nil)),
		"BindUnmarshaler": reflect.ValueOf((*echo.BindUnmarshaler)(nil)),
		"Binder": reflect.ValueOf((*echo.Binder)(nil)),
		"BindingError": reflect.ValueOf((*echo.BindingError)(nil)),
		"Config": reflect.ValueOf((*echo.Config)(nil)),
		"Context": reflect.ValueOf((*echo.Context)(nil)),
		"DefaultBinder": reflect.ValueOf((*echo.DefaultBinder)(nil)),
		"DefaultJSONSerializer": reflect.ValueOf((*echo.DefaultJSONSerializer)(nil)),
		"DefaultRouter": reflect.ValueOf((*echo.DefaultRouter)(nil)),
		"Echo": reflect.ValueOf((*echo.Echo)(nil)),
		"Group": reflect.ValueOf((*echo.Group)(nil)),
		"HTTPError": reflect.ValueOf((*echo.HTTPError)(nil)),
		"HTTPErrorHandler": reflect.ValueOf((*echo.HTTPErrorHandler)(nil)),
		"HTTPStatusCoder": reflect.ValueOf((*echo.HTTPStatusCoder)(nil)),
		"HandlerFunc": reflect.ValueOf((*echo.HandlerFunc)(nil)),
		"IPExtractor": reflect.ValueOf((*echo.IPExtractor)(nil)),
		"JSONSerializer": reflect.ValueOf((*echo.JSONSerializer)(nil)),
		"MiddlewareConfigurator": reflect.ValueOf((*echo.MiddlewareConfigurator)(nil)),
		"MiddlewareFunc": reflect.ValueOf((*echo.MiddlewareFunc)(nil)),
		"PathValue": reflect.ValueOf((*echo.PathValue)(nil)),
		"PathValues": reflect.ValueOf((*echo.PathValues)(nil)),
		"Renderer": reflect.ValueOf((*echo.Renderer)(nil)),
		"Response": reflect.ValueOf((*echo.Response)(nil)),
		"Route": reflect.ValueOf((*echo.Route)(nil)),
		"RouteInfo": reflect.ValueOf((*echo.RouteInfo)(nil)),
		"Router": reflect.ValueOf((*echo.Router)(nil)),
		"RouterConfig": reflect.ValueOf((*echo.RouterConfig)(nil)),
		"Routes": reflect.ValueOf((*echo.Routes)(nil)),
		"StartConfig": reflect.ValueOf((*echo.StartConfig)(nil)),
		"TemplateRenderer": reflect.ValueOf((*echo.TemplateRenderer)(nil)),
		"TimeLayout": reflect.ValueOf((*echo.TimeLayout)(nil)),
		"TimeOpts": reflect.ValueOf((*echo.TimeOpts)(nil)),
		"TrustOption": reflect.ValueOf((*echo.TrustOption)(nil)),
		"Validator": reflect.ValueOf((*echo.Validator)(nil)),
		"ValueBinder": reflect.ValueOf((*echo.ValueBinder)(nil)),
		
		// interface wrapper definitions
		"_BindUnmarshaler": reflect.ValueOf((*_github_com_labstack_echo_v5_BindUnmarshaler)(nil)),
		"_Binder": reflect.ValueOf((*_github_com_labstack_echo_v5_Binder)(nil)),
		"_HTTPStatusCoder": reflect.ValueOf((*_github_com_labstack_echo_v5_HTTPStatusCoder)(nil)),
		"_JSONSerializer": reflect.ValueOf((*_github_com_labstack_echo_v5_JSONSerializer)(nil)),
		"_MiddlewareConfigurator": reflect.ValueOf((*_github_com_labstack_echo_v5_MiddlewareConfigurator)(nil)),
		"_Renderer": reflect.ValueOf((*_github_com_labstack_echo_v5_Renderer)(nil)),
		"_Router": reflect.ValueOf((*_github_com_labstack_echo_v5_Router)(nil)),
		"_Validator": reflect.ValueOf((*_github_com_labstack_echo_v5_Validator)(nil)),
		
	}
}
// _github_com_labstack_echo_v5_BindUnmarshaler is an interface wrapper for BindUnmarshaler type
	type _github_com_labstack_echo_v5_BindUnmarshaler struct {
		IValue interface{}
		WUnmarshalParam func(param string) ( error)
		
	}
	func (W _github_com_labstack_echo_v5_BindUnmarshaler) UnmarshalParam(param string) ( error) {return W.WUnmarshalParam(param)
		}
	
// _github_com_labstack_echo_v5_Binder is an interface wrapper for Binder type
	type _github_com_labstack_echo_v5_Binder struct {
		IValue interface{}
		WBind func(c *echo.Context, target any) ( error)
		
	}
	func (W _github_com_labstack_echo_v5_Binder) Bind(c *echo.Context, target any) ( error) {return W.WBind(c, target)
		}
	
// _github_com_labstack_echo_v5_HTTPStatusCoder is an interface wrapper for HTTPStatusCoder type
	type _github_com_labstack_echo_v5_HTTPStatusCoder struct {
		IValue interface{}
		WStatusCode func() ( int)
		
	}
	func (W _github_com_labstack_echo_v5_HTTPStatusCoder) StatusCode() ( int) {return W.WStatusCode()
		}
	
// _github_com_labstack_echo_v5_JSONSerializer is an interface wrapper for JSONSerializer type
	type _github_com_labstack_echo_v5_JSONSerializer struct {
		IValue interface{}
		WDeserialize func(c *echo.Context, target any) ( error)
		WSerialize func(c *echo.Context, target any, indent string) ( error)
		
	}
	func (W _github_com_labstack_echo_v5_JSONSerializer) Deserialize(c *echo.Context, target any) ( error) {return W.WDeserialize(c, target)
		}
	func (W _github_com_labstack_echo_v5_JSONSerializer) Serialize(c *echo.Context, target any, indent string) ( error) {return W.WSerialize(c, target, indent)
		}
	
// _github_com_labstack_echo_v5_MiddlewareConfigurator is an interface wrapper for MiddlewareConfigurator type
	type _github_com_labstack_echo_v5_MiddlewareConfigurator struct {
		IValue interface{}
		WToMiddleware func() ( echo.MiddlewareFunc,  error)
		
	}
	func (W _github_com_labstack_echo_v5_MiddlewareConfigurator) ToMiddleware() ( echo.MiddlewareFunc,  error) {return W.WToMiddleware()
		}
	
// _github_com_labstack_echo_v5_Renderer is an interface wrapper for Renderer type
	type _github_com_labstack_echo_v5_Renderer struct {
		IValue interface{}
		WRender func(c *echo.Context, w io.Writer, templateName string, data any) ( error)
		
	}
	func (W _github_com_labstack_echo_v5_Renderer) Render(c *echo.Context, w io.Writer, templateName string, data any) ( error) {return W.WRender(c, w, templateName, data)
		}
	
// _github_com_labstack_echo_v5_Router is an interface wrapper for Router type
	type _github_com_labstack_echo_v5_Router struct {
		IValue interface{}
		WAdd func(routable echo.Route) ( echo.RouteInfo,  error)
		WRemove func(method string, path string) ( error)
		WRoute func(c *echo.Context) ( echo.HandlerFunc)
		WRoutes func() ( echo.Routes)
		
	}
	func (W _github_com_labstack_echo_v5_Router) Add(routable echo.Route) ( echo.RouteInfo,  error) {return W.WAdd(routable)
		}
	func (W _github_com_labstack_echo_v5_Router) Remove(method string, path string) ( error) {return W.WRemove(method, path)
		}
	func (W _github_com_labstack_echo_v5_Router) Route(c *echo.Context) ( echo.HandlerFunc) {return W.WRoute(c)
		}
	func (W _github_com_labstack_echo_v5_Router) Routes() ( echo.Routes) {return W.WRoutes()
		}
	
// _github_com_labstack_echo_v5_Validator is an interface wrapper for Validator type
	type _github_com_labstack_echo_v5_Validator struct {
		IValue interface{}
		WValidate func(i any) ( error)
		
	}
	func (W _github_com_labstack_echo_v5_Validator) Validate(i any) ( error) {return W.WValidate(i)
		}
	


