import {test, expect} from '../../support/fixtures'

test('runtime configuration gates provider links and operations', async ({authenticatedPage: page, apiContext, userToken}) => {
	const oauth = ['todoist', 'trello', 'microsoft-todo']
	const files = ['csv', 'ticktick', 'wekan', 'vikunja-file', 'planka']
	const info = await (await apiContext.get('/api/v2/info')).json()
	await page.goto('/user/settings/migrate')
	for (const provider of [...files, ...oauth]) {
		const available = info.available_migrators.includes(provider)
		if (files.includes(provider)) expect(available).toBe(true)
		await expect(page.locator(`a.migration-service-link[href="/migrate/${provider}"]`)).toHaveCount(available ? 1 : 0)
		const status = await apiContext.get(`/api/v2/migration/${provider}/status`, {headers: {Authorization: `Bearer ${userToken}`}})
		expect(status.status()).toBe(available ? 200 : 404)
		if (oauth.includes(provider)) {
			const auth = await apiContext.get(`/api/v2/migration/${provider}/auth`, {headers: {Authorization: `Bearer ${userToken}`}})
			expect(auth.status()).toBe(available ? 200 : 404)
			if (available) expect((await auth.json()).url).toMatch(/^https:\/\//)
		}
	}
})
