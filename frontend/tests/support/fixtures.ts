import {test as base, type APIRequestContext, type Page} from '@playwright/test'
import {Factory} from './factory'
import {login, createFakeUser} from './authenticateUser'
import type {UserAttributes} from '../factories/user'

export const test = base.extend<{
	apiContext: APIRequestContext;
	authenticatedPage: Page;
	currentUser: UserAttributes;
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

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	currentUser: async ({apiContext}, use) => {
		const user = await createFakeUser()
		await use(user)
	},

	authenticatedPage: async ({page, apiContext, currentUser}, use) => {
		await login(page, apiContext, currentUser)
		await use(page)
	},
})

export {expect} from '@playwright/test'
