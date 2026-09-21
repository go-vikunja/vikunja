import {test, expect} from '../../support/fixtures'
import {NotificationFactory} from '../../factories/notification'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

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
