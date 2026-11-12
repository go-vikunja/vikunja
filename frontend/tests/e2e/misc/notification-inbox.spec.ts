import {test, expect} from '../../support/fixtures'
import {NotificationFactory} from '../../factories/notification'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

function isInboxRequest(url: string) {
	return new URL(url).pathname.endsWith('/api/v2/notifications')
}

test('notification read and clear state survives reload', async ({authenticatedPage: page, currentUser, apiContext, userToken}) => {
	const [project] = await ProjectFactory.create(1, {owner_id: currentUser.id})
	const [task] = await TaskFactory.create(1, {project_id: project.id, created_by_id: currentUser.id})
	await NotificationFactory.create(2, {
		notifiable_id: currentUser.id,
		project_id: project.id,
		notification: JSON.stringify({
			doer: currentUser,
			task,
		}),
	})
	await page.goto('/')
	const open = () => page.locator('.notifications .trigger-button').click()
	const rows = page.locator('.single-notification')
	await open()
	await expect(rows).toHaveCount(2)
	await rows.first().click()
	await open()
	await expect(page.locator('.single-notification .read-indicator.read')).toHaveCount(1)
	await page.reload()
	await open()
	await expect(page.locator('.single-notification .read-indicator.read')).toHaveCount(1)
	await page.getByRole('button', {name: 'Mark all notifications as read'}).click()
	await expect(page.locator('.single-notification .read-indicator.read')).toHaveCount(2)
	await page.reload()
	await open()
	await expect(page.locator('.single-notification .read-indicator.read')).toHaveCount(2)
	await page.getByRole('button', {name: 'Clear notifications', exact: true}).click()
	await expect(rows).toHaveCount(0)
	await page.reload()
	await open()
	await expect(rows).toHaveCount(0)
	const response = await apiContext.get('/api/v2/notifications', {headers: {Authorization: `Bearer ${userToken}`}})
	expect(response.ok()).toBe(true)
	const body = await response.json()
	expect(body.items).toEqual([])
	expect(body.total).toBe(0)
})

test('the inbox loads older notifications one page at a time', async ({authenticatedPage: page, currentUser}) => {
	const [project] = await ProjectFactory.create(1, {owner_id: currentUser.id})
	const [task] = await TaskFactory.create(1, {project_id: project.id, created_by_id: currentUser.id})
	await NotificationFactory.create(21, {
		notifiable_id: currentUser.id,
		project_id: project.id,
		notification: JSON.stringify({
			doer: currentUser,
			task,
		}),
	})
	const pageRequests: string[] = []
	page.on('request', request => {
		if (isInboxRequest(request.url())) pageRequests.push(new URL(request.url()).search)
	})
	await page.goto('/')
	await page.locator('.notifications .trigger-button').click()
	const rows = page.locator('.single-notification')
	const loadMore = page.getByRole('button', {name: 'Load older notifications'})
	await expect(rows).toHaveCount(20)
	await loadMore.click()
	await expect(rows).toHaveCount(21)
	await expect(loadMore).toHaveCount(0)
	expect(pageRequests).toContain('?page=1&per_page=20')
	expect(pageRequests).toContain('?page=2&per_page=20')
})

test('fetches the inbox once per page load', async ({authenticatedPage: page}) => {
	const events: string[] = []
	// Delay the socket so the inbox response beats auth.success, the old refetch trigger.
	await page.routeWebSocket('**/api/v2/ws', ws => {
		const server = ws.connectToServer()
		server.onMessage(message => {
			setTimeout(() => {
				if (message.toString().includes('"auth.success"')) events.push('auth.success')
				ws.send(message)
			}, 300)
		})
	})
	page.on('request', request => {
		if (isInboxRequest(request.url())) events.push('inbox')
	})
	await page.goto('/')
	await expect.poll(() => events).toEqual(['auth.success', 'inbox'])
	await page.waitForTimeout(1000)
	expect(events).toEqual(['auth.success', 'inbox'])
})

test('keeps polling the inbox when the websocket is unavailable', async ({authenticatedPage: page}) => {
	await page.routeWebSocket('**/api/v2/ws', ws => ws.close())
	let inboxRequests = 0
	page.on('request', request => {
		if (isInboxRequest(request.url())) inboxRequests++
	})
	await page.goto('/')
	await expect.poll(() => inboxRequests).toBe(1)
	await expect.poll(() => inboxRequests, {timeout: 13_000}).toBe(2) // 10s refetchInterval + slack
})
