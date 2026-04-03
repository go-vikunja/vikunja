import {test as base, type APIRequestContext, type Page} from '@playwright/test'
import {Factory} from './factory'
import {login, createFakeUser} from './authenticateUser'

export const test = base.extend<{
	apiContext: APIRequestContext;
	authenticatedPage: Page;
	currentUser: any;
	userToken: string;
}>({
	apiContext: async ({playwright}, use) => {
		const baseURL = process.env.API_URL || 'http://localhost:3456/api/v1/'
		const apiContext = await playwright.request.newContext({
			baseURL,
		})

		Factory.setRequestContext(apiContext)
		await use(apiContext)
		await apiContext.dispose()
	},

	currentUser: async ({apiContext}, use) => {
		await Factory.truncateAll()
		const user = await createFakeUser()
		await use(user)
	},

	userToken: async ({apiContext, currentUser}, use) => {
		const {token} = await login(null, apiContext, currentUser)
		await use(token)
	},

	authenticatedPage: async ({page, apiContext, currentUser}, use) => {
		const {token} = await login(page, apiContext, currentUser)
		await use(page)
		// Navigate away to stop all frontend requests (notification polling, token
		// refresh, etc.) before the next test's fixture setup seeds the database.
		// Without this, the previous test's page can hold DB connections via API
		// requests, starving the next test's Factory.seed() PATCH call.
		await page.goto('about:blank').catch(() => {})
	},
})

export {expect} from '@playwright/test'
