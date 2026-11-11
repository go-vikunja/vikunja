import type {Page} from '@playwright/test'
import {
	test,
	expect,
} from '../../support/fixtures'
import {LicenseFactory} from '../../factories/license'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {TEST_PASSWORD} from '../../support/constants'

// Subscriptions are sent in one batch after auth, so the last event of the set proves all of them arrived.
const LAST_SUBSCRIBED_EVENT = 'notification.created'

function waitForSubscription(page: Page, event: string) {
	return page.waitForEvent('websocket').then(socket => socket.waitForEvent('framesent', {
		predicate: frame => {
			const message = JSON.parse(frame.payload.toString())
			return message.action === 'subscribe' && message.event === event
		},
	}))
}

test('comment notifications refresh the open task without a reload', async ({
	authenticatedPage: page,
	apiContext,
	currentUser,
	userToken,
}) => {
	const [commenter] = await UserFactory.create(1, {id: 100}, false)
	const login = await apiContext.post('login', {data: {
		username: commenter.username,
		password: TEST_PASSWORD,
	}})
	const {token} = await login.json()
	await ProjectFactory.create(1, {
		id: 100,
		owner_id: 100,
	}, false)
	await ProjectViewFactory.create(1, {
		id: 100,
		project_id: 100,
	}, false)
	await TaskFactory.create(1, {
		id: 100,
		project_id: 100,
		created_by_id: 100,
	}, false)
	await UserProjectFactory.create(1, {
		id: 100,
		project_id: 100,
		user_id: currentUser.id,
	}, false)
	const subscribed = waitForSubscription(page, LAST_SUBSCRIBED_EVENT)
	await page.goto('/tasks/100')
	await subscribed
	await expect(page.locator('.comments .comment')).toHaveCount(0)
	const comment = `<p>Live update <mention-user data-id="${currentUser.username}">@${currentUser.username}</mention-user></p>`
	const response = await apiContext.put('tasks/100/comments', {
		data: {comment},
		headers: {Authorization: `Bearer ${token}`},
	})
	expect(response.ok()).toBeTruthy()
	await expect(page.locator('.comments .comment').filter({hasText: 'Live update'})).toBeVisible()
	await page.reload()
	await expect(page.locator('.comments .comment').filter({hasText: 'Live update'})).toBeVisible()
	const stored = await apiContext.get('/api/v2/tasks/100/comments', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await stored.json()).items).toEqual([expect.objectContaining({comment})])
})

test.describe('timer cache events', () => {
	test.beforeEach(async () => { await LicenseFactory.enable(['time_tracking']) })
	test.afterEach(async () => { await LicenseFactory.disable() })

	test('another client starts, stops and deletes a timer in the open list', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		const [project] = await ProjectFactory.create(1, {title: 'Remote timer project'}, false)
		const subscribed = waitForSubscription(page, LAST_SUBSCRIBED_EVENT)
		await page.goto('/time-tracking')
		await subscribed
		await expect(page.locator('[data-cy="addTimeEntry"]')).toBeVisible()
		const headers = {Authorization: `Bearer ${userToken}`}
		const created = await apiContext.post('/api/v2/time-entries', {
			headers,
			data: {
				project_id: project.id,
				start_time: new Date().toISOString(),
				comment: 'Remote timer',
			},
		})
		expect(created.ok()).toBeTruthy()
		const entry = await created.json()
		const badge = page.locator('[data-cy="timerBadge"]')
		const rows = page.locator('[data-cy="timeEntry"]')
		await expect(badge).toBeVisible()
		await expect(rows).toHaveCount(1)
		await expect(rows).toContainText('Remote timer')
		const stopped = await apiContext.post('/api/v2/time-entries/timer/stop', {headers})
		expect(stopped.ok()).toBeTruthy()
		await expect(badge).not.toBeVisible()
		await expect(rows).toHaveCount(1)
		// A running entry renders an open-ended `start – …`.
		await expect(rows).not.toContainText('…')
		const resubscribed = waitForSubscription(page, LAST_SUBSCRIBED_EVENT)
		await page.reload()
		await resubscribed
		await expect(rows).toHaveCount(1)
		await expect(rows).not.toContainText('…')
		const deleted = await apiContext.delete(`/api/v2/time-entries/${entry.id}`, {headers})
		expect(deleted.ok()).toBeTruthy()
		await expect(rows).toHaveCount(0)
		const stored = await apiContext.get('/api/v2/time-entries', {headers})
		expect((await stored.json()).items ?? []).toEqual([])
	})
})
