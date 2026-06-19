const crypto = require('crypto')
const {net} = require('electron')

const CLIENT_ID = 'vikunja-desktop'
const REDIRECT_URI = 'vikunja-desktop://callback'

let pendingCodeVerifier = null

function generateCodeVerifier() {
	return crypto.randomBytes(32).toString('base64url')
}

function generateCodeChallenge(verifier) {
	return crypto.createHash('sha256').update(verifier).digest('base64url')
}

function buildAuthorizationUrl(frontendUrl, codeChallenge) {
	// Strip trailing slash and /api/v1 suffix to get the frontend origin
	let base = frontendUrl.replace(/\/+$/, '').replace(/\/api\/v1$/, '')

	const url = new URL(base)
	url.pathname = url.pathname.replace(/\/+$/, '') + '/oauth/authorize'
	url.searchParams.set('response_type', 'code')
	url.searchParams.set('client_id', CLIENT_ID)
	url.searchParams.set('redirect_uri', REDIRECT_URI)
	url.searchParams.set('code_challenge', codeChallenge)
	url.searchParams.set('code_challenge_method', 'S256')

	return url.toString()
}

function startLogin(apiUrl) {
	const verifier = generateCodeVerifier()
	const challenge = generateCodeChallenge(verifier)
	pendingCodeVerifier = verifier

	return buildAuthorizationUrl(apiUrl, challenge)
}

function postJSON(url, body) {
	return new Promise((resolve, reject) => {
		const request = net.request({
			method: 'POST',
			url,
		})
		request.setHeader('Content-Type', 'application/json')

		let responseData = ''

		request.on('response', (response) => {
			response.on('data', (chunk) => {
				responseData += chunk.toString()
			})
			response.on('end', () => {
				try {
					const parsed = JSON.parse(responseData)
					if (response.statusCode >= 200 && response.statusCode < 300) {
						resolve(parsed)
					} else {
						reject(new Error(parsed.message || `HTTP ${response.statusCode}`))
					}
				} catch {
					reject(new Error(`Invalid JSON response: ${responseData.substring(0, 200)}`))
				}
			})
		})

		request.on('error', (err) => {
			reject(err)
		})

		request.write(JSON.stringify(body))
		request.end()
	})
}

function getTokenEndpoint(apiUrl) {
	let base = apiUrl.replace(/\/+$/, '')
	if (!base.endsWith('/api/v1')) {
		base += '/api/v1'
	}
	return `${base}/oauth/token`
}

async function exchangeCodeForTokens(apiUrl, code) {
	const verifier = pendingCodeVerifier
	pendingCodeVerifier = null

	if (!verifier) {
		throw new Error('No pending PKCE verifier found')
	}

	const tokenUrl = getTokenEndpoint(apiUrl)
	return postJSON(tokenUrl, {
		grant_type: 'authorization_code',
		code,
		client_id: CLIENT_ID,
		redirect_uri: REDIRECT_URI,
		code_verifier: verifier,
	})
}

async function refreshAccessToken(apiUrl, refreshToken) {
	const tokenUrl = getTokenEndpoint(apiUrl)
	return postJSON(tokenUrl, {
		grant_type: 'refresh_token',
		refresh_token: refreshToken,
	})
}

module.exports = {
	startLogin,
	exchangeCodeForTokens,
	refreshAccessToken,
}
