import {createHash, randomBytes} from 'crypto'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

test.describe('OAuth 2.0 Authorization Flow', () => {
	let username: string

	test.beforeEach(async ({apiContext}) => {
		const [user] = await UserFactory.create(1)
		username = user.username
	})

	test('Full browser authorization code flow with PKCE', async ({page, apiContext}) => {
		await setupApiUrl(page)

		// Generate PKCE code_verifier and code_challenge (S256)
		const codeVerifier = randomBytes(32).toString('base64url')
		const codeChallenge = createHash('sha256').update(codeVerifier).digest('base64url')
		const state = randomBytes(16).toString('base64url')

		// Build the authorize URL as a frontend route with OAuth query params.
		// The OAuthAuthorize.vue component reads these and POSTs to the API.
		const authorizeParams = new URLSearchParams({
			response_type: 'code',
			client_id: 'vikunja',
			redirect_uri: 'vikunja-flutter://callback',
			code_challenge: codeChallenge,
			code_challenge_method: 'S256',
			state,
		})

		// Navigate to the OAuth authorize frontend route.
		// The user is not logged in, so the router guard redirects to /login while
		// carrying the authorize destination in a copyable #redirect= hash (not a
		// query param, to keep the OAuth params out of access logs).
		await page.goto(`/oauth/authorize?${authorizeParams}`)
		await expect(page).toHaveURL(/\/login#redirect=/)

		// The decoded #redirect= destination must carry the full authorize URL, including the
		// OAuth params — checking only for the path would pass even if the query were dropped.
		const redirectHash = decodeURIComponent(new URL(page.url()).hash)
		expect(redirectHash).toContain('/oauth/authorize')
		expect(redirectHash).toContain('response_type=code')
		expect(redirectHash).toContain('client_id=vikunja')
		expect(redirectHash).toContain(`code_challenge=${codeChallenge}`)
		expect(redirectHash).toContain(`state=${state}`)

		// Register the response listener BEFORE clicking Login, because after
		// login redirectIfSaved() navigates back to /oauth/authorize and the
		// component immediately POSTs to the API.
		const authorizeResponsePromise = page.waitForResponse(
			response => response.url().includes('/api/v1/oauth/authorize') && response.request().method() === 'POST',
			{timeout: 15000},
		)

		// Log in via the browser UI
		await page.locator('input[id=username]').fill(username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		// Wait for the authorize API call that fires after login redirect
		const authorizeResponse = await authorizeResponsePromise
		const authorizeBody = await authorizeResponse.json()
		expect(authorizeBody.code).toBeTruthy()
		expect(authorizeBody.redirect_uri).toBe('vikunja-flutter://callback')
		expect(authorizeBody.state).toBe(state)

		const code = authorizeBody.code

		// Exchange the authorization code for tokens
		const tokenResponse = await apiContext.post('oauth/token', {
			data: {
				grant_type: 'authorization_code',
				code,
				client_id: 'vikunja',
				redirect_uri: 'vikunja-flutter://callback',
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

	// The primary #2654 scenario: the native client opened a different default browser that is
	// already signed in to Vikunja. Opening the copied /login#redirect=<oauth.authorize> URL must
	// run the OAuth flow with the existing session instead of short-circuiting to home.
	test('Already-authenticated browser opening the copied login redirect runs the authorize flow', async ({authenticatedPage, apiContext, currentUser}) => {
		const page = authenticatedPage

		const codeVerifier = randomBytes(32).toString('base64url')
		const codeChallenge = createHash('sha256').update(codeVerifier).digest('base64url')
		const state = randomBytes(16).toString('base64url')

		const authorizeParams = new URLSearchParams({
			response_type: 'code',
			client_id: 'vikunja',
			redirect_uri: 'vikunja-flutter://callback',
			code_challenge: codeChallenge,
			code_challenge_method: 'S256',
			state,
		})

		// The component POSTs as soon as it mounts with the existing session, so register the
		// listener before navigating.
		const authorizeResponsePromise = page.waitForResponse(
			response => response.url().includes('/api/v1/oauth/authorize') && response.request().method() === 'POST',
			{timeout: 15000},
		)

		// Open the copyable login URL exactly as it would be pasted from another browser
		// (#redirect= is REDIRECT_HASH_PREFIX from @/constants/redirectHash, inlined here because
		// the e2e runner has no @ alias).
		const redirectDestination = `/oauth/authorize?${authorizeParams}`
		await page.goto(`/login#redirect=${encodeURIComponent(redirectDestination)}`)

		// The authed guard must send us straight to /oauth/authorize, not home.
		await expect(page).toHaveURL(/\/oauth\/authorize/)
		const landed = new URL(page.url())
		expect(landed.pathname).toBe('/oauth/authorize')
		expect(landed.searchParams.get('response_type')).toBe('code')
		expect(landed.searchParams.get('client_id')).toBe('vikunja')
		expect(landed.searchParams.get('code_challenge')).toBe(codeChallenge)
		expect(landed.searchParams.get('state')).toBe(state)

		// The PKCE flow completes with the existing session — no second login.
		const authorizeResponse = await authorizeResponsePromise
		const authorizeBody = await authorizeResponse.json()
		expect(authorizeBody.code).toBeTruthy()
		expect(authorizeBody.redirect_uri).toBe('vikunja-flutter://callback')
		expect(authorizeBody.state).toBe(state)

		const tokenResponse = await apiContext.post('oauth/token', {
			data: {
				grant_type: 'authorization_code',
				code: authorizeBody.code,
				client_id: 'vikunja',
				redirect_uri: 'vikunja-flutter://callback',
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
