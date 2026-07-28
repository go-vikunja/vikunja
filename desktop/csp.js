const crypto = require('crypto')
const fs = require('fs')
const path = require('path')

// Matches <script> tags without a src attribute, i.e. those carrying inline code.
const INLINE_SCRIPT_RE = /<script\b(?![^>]*\ssrc=)[^>]*>([\s\S]*?)<\/script>/gi

// build.js rewrites the inline script that sets window.API_URL, so hash what's on disk.
function inlineScriptHashes(frontendDir) {
	const html = fs.readFileSync(path.join(frontendDir, 'index.html'), 'utf8')
	return [...html.matchAll(INLINE_SCRIPT_RE)]
		// The HTML parser normalizes CRLF and lone CR to LF before the hash is taken.
		.map(([, body]) => `'sha256-${crypto.createHash('sha256').update(body.replace(/\r\n?/g, '\n'), 'utf8').digest('base64')}'`)
}

function buildContentSecurityPolicy(frontendDir) {
	let scriptSrc = "'self'"
	try {
		const hashes = inlineScriptHashes(frontendDir)
		if (hashes.length === 0) {
			throw new Error('no inline script found in index.html')
		}
		scriptSrc += ' ' + hashes.join(' ')
	} catch (err) {
		// Blocking that script leaves window.API_URL undefined, which throws before the app
		// mounts. A weaker policy beats a blank window.
		console.warn('Could not hash inline scripts, falling back to script-src unsafe-inline:', err.message)
		scriptSrc += " 'unsafe-inline'"
	}

	return [
		"default-src 'self'",
		`script-src ${scriptSrc}`,
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
		// registerServiceWorker.ts registers sw.js.
		"worker-src 'self'",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	].join('; ')
}

module.exports = {
	buildContentSecurityPolicy,
}
