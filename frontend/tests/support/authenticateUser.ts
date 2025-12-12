import type {Page, APIRequestContext} from '@playwright/test'
import {UserFactory} from '../factories/user'
import {TEST_PASSWORD} from './constants'

/**
 * This authenticates a user and puts the token in local storage which allows us to perform authenticated requests.
 * Returns the user and token for use in tests that need to make authenticated API calls.
 */
export async function login(page: Page | null, apiContext: APIRequestContext, user?: any) {
	if (!user) {
		throw new Error('Needs user')
	}

	// Login via API
	const response = await apiContext.post('login', {
		data: {
			username: user.username,
			password: TEST_PASSWORD,
		},
	})

	if (!response.ok()) {
		throw new Error(`Login failed: ${response.status()} ${response.statusText()}`)
	}

	const body = await response.json()
	const token = body.token

	// Set token and API_URL in localStorage before navigating (only if page is provided)
	if (page) {
		// Use 127.0.0.1 instead of localhost to match the frontend's origin for CORS
		const apiUrl = process.env.API_URL || 'http://127.0.0.1:3456/api/v1'
		await page.addInitScript(({token, apiUrl}) => {
			window.localStorage.setItem('token', token)
			window.localStorage.setItem('API_URL', apiUrl)
		}, {token, apiUrl})
	}

	return {user, token}
}

export async function createFakeUser() {
	const [u] = await UserFactory.create(1)
	return u
}

/**
 * Helper function to set up authentication for a test suite
 * Returns the created user for use in tests
 */
export function createFakeUserAndLogin() {
	// This returns undefined and instead relies on Playwright's beforeEach hooks
	// The actual user will be available through the test context
	return undefined
}
