const CONTENT_SECURITY_POLICY = [
	"default-src 'self'",
	// build.js externalises the inline window.API_URL script into /api-url.js, so no hash
	// or 'unsafe-inline' is needed here.
	"script-src 'self'",
	// TipTap, vue3-notification and the emoji picker append <style> elements at runtime.
	"style-src 'self' 'unsafe-inline'",
	// Task descriptions embed remote images; data: for bundled icons, blob: for avatars,
	// project backgrounds and the TOTP QR code.
	'img-src * data: blob:',
	"font-src 'self'",
	// The API URL is user-configurable to any remote instance and the realtime socket URL
	// is derived from it, so no origin allowlist is possible.
	'connect-src * ws: wss:',
	// PDF attachment previews render a blob: URL in an iframe.
	'frame-src blob:',
	"worker-src 'self'",
	"object-src 'none'",
	"base-uri 'self'",
	"form-action 'self'",
	"frame-ancestors 'none'",
].join('; ')

module.exports = {CONTENT_SECURITY_POLICY}
