import {createHash, randomBytes} from 'crypto'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TEST_PASSWORD} from '../../support/constants'

test.describe('OAuth 2.0 Authorization Flow', () => {
	let username: string

	test.beforeEach(async ({apiContext}) => {
		const [user] = await UserFactory.create(1)
		username = user.username
	})

	test('Full browser authorization code flow with PKCE', async ({page, apiContext}) => {
		const apiUrl = (process.env.API_URL || 'http://127.0.0.1:3456/api/v1').replace(/\/+$/, '')
		const frontendBase = (process.env.BASE_URL || 'http://127.0.0.1:4173').replace(/\/+$/, '')

		// Set the API URL so the frontend knows where to send requests
		await page.addInitScript(({apiUrl}) => {
			window.localStorage.setItem('API_URL', apiUrl)
			window.API_URL = apiUrl
		}, {apiUrl})

		// Generate PKCE code_verifier and code_challenge (S256)
		const codeVerifier = randomBytes(32).toString('base64url')
		const codeChallenge = createHash('sha256').update(codeVerifier).digest('base64url')
		const state = randomBytes(16).toString('base64url')

		// Build the authorize URL on the frontend's origin so the same-origin
		// check in Login.vue passes when it reads the redirect query param.
		// In production the API and frontend share an origin; in the E2E test
		// they run on different ports, so we route-intercept the request below.
		const authorizeParams = new URLSearchParams({
			response_type: 'code',
			client_id: 'vikunja-flutter',
			redirect_uri: 'vikunja://callback',
			code_challenge: codeChallenge,
			code_challenge_method: 'S256',
			state,
		})
		const authorizeUrl = `${frontendBase}/api/v1/oauth/authorize?${authorizeParams}`

		// Capture the JWT from the login API response so the route handler
		// can forward it to the real authorize endpoint.
		let jwt = ''
		const jwtReady = new Promise<void>(resolve => {
			page.on('response', async response => {
				if (
					response.url().includes('/login') &&
					response.request().method() === 'POST' &&
					response.ok()
				) {
					try {
						const body = await response.json()
						jwt = body.token
					} catch { /* ignore parse errors */ }
					resolve()
				}
			})
		})

		// Intercept authorize requests on the frontend origin and proxy them
		// to the real API server with the JWT Authorization header.
		// This is necessary because the E2E test runs the API and frontend
		// on separate ports, while in production they share an origin.
		let capturedLocation = ''
		let resolveAuthorize: () => void
		const authorizeHandled = new Promise<void>(resolve => {
			resolveAuthorize = resolve
		})

		await page.route('**/api/v1/oauth/authorize**', async route => {
			// Wait for the JWT to be available from the login response
			await jwtReady

			const requestUrl = new URL(route.request().url())
			const apiResponse = await apiContext.get(
				`${apiUrl}/oauth/authorize${requestUrl.search}`,
				{
					headers: {'Authorization': `Bearer ${jwt}`},
					maxRedirects: 0,
				},
			)

			capturedLocation = apiResponse.headers()['location'] || ''

			await route.fulfill({
				status: apiResponse.status(),
				headers: apiResponse.headers(),
			})

			resolveAuthorize()
		})

		// Navigate to the frontend login page with the OAuth redirect parameter
		await page.goto(`/login?redirect=${encodeURIComponent(authorizeUrl)}`)
		await expect(page).toHaveURL(/\/login/)

		// Log in via the browser UI
		await page.locator('input[id=username]').fill(username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		// After login, Login.vue reads route.query.redirect, validates
		// same-origin, and does window.location.href = authorizeURL.
		// The route handler intercepts this, proxies to the real API,
		// and the API responds with 302 → vikunja://callback?code=...
		await authorizeHandled

		// Verify the API returned a redirect to vikunja://callback with code and state
		expect(capturedLocation).toContain('vikunja://callback')
		const callbackUrl = new URL(capturedLocation)
		const code = callbackUrl.searchParams.get('code')
		expect(code).toBeTruthy()
		expect(callbackUrl.searchParams.get('state')).toBe(state)

		// Exchange the authorization code for tokens
		const tokenResponse = await apiContext.post('oauth/token', {
			form: {
				grant_type: 'authorization_code',
				code: code!,
				client_id: 'vikunja-flutter',
				redirect_uri: 'vikunja://callback',
				code_verifier: codeVerifier,
			},
		})

		expect(tokenResponse.ok()).toBe(true)
		const tokenBody = await tokenResponse.json()
		expect(tokenBody.access_token).toBeTruthy()
		expect(tokenBody.refresh_token).toBeTruthy()
		expect(tokenBody.token_type).toBe('bearer')
		expect(tokenBody.expires_in).toBeGreaterThan(0)
	})
})
