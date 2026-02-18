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
	},
})

export {expect} from '@playwright/test'
