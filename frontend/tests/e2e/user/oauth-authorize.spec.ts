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
		// The user is not logged in, so the router guard saves the route
		// and redirects to /login.
		await page.goto(`/oauth/authorize?${authorizeParams}`)
		await expect(page).toHaveURL(/\/login/)

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
})
