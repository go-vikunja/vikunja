import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TEST_PASSWORD} from '../../support/constants'

test.describe('WebSocket Frontend Integration', () => {

	test('notification badge updates in real-time when added to team', async ({
		authenticatedPage: page,
		apiContext,
		currentUser,
	}) => {
		// Navigate to the app so WebSocket connects
		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Verify no unread indicator initially
		await expect(page.locator('.notifications .unread-indicator')).toHaveCount(0)

		// Create a second user who will add currentUser to a team
		const [userA] = await UserFactory.create(1, {id: 100}, false)
		const loginResponse = await apiContext.post('login', {
			data: {username: userA.username, password: TEST_PASSWORD},
		})
		const {token: tokenA} = await loginResponse.json()

		// User A creates a team
		const teamResponse = await apiContext.put('teams', {
			data: {name: 'Real-Time Test Team'},
			headers: {Authorization: `Bearer ${tokenA}`},
		})
		const team = await teamResponse.json()

		// User A adds currentUser to the team — this triggers a notification
		await apiContext.put(`teams/${team.id}/members`, {
			data: {username: currentUser.username},
			headers: {Authorization: `Bearer ${tokenA}`},
		})

		// The unread indicator should appear without page refresh
		await expect(page.locator('.notifications .unread-indicator')).toBeVisible({
			timeout: 10000,
		})
	})

	test('notification appears in dropdown after real-time delivery', async ({
		authenticatedPage: page,
		apiContext,
		currentUser,
	}) => {
		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Create user A and trigger notification
		const [userA] = await UserFactory.create(1, {id: 100}, false)
		const loginResponse = await apiContext.post('login', {
			data: {username: userA.username, password: TEST_PASSWORD},
		})
		const {token: tokenA} = await loginResponse.json()

		const teamResponse = await apiContext.put('teams', {
			data: {name: 'Dropdown Test Team'},
			headers: {Authorization: `Bearer ${tokenA}`},
		})
		const team = await teamResponse.json()

		await apiContext.put(`teams/${team.id}/members`, {
			data: {username: currentUser.username},
			headers: {Authorization: `Bearer ${tokenA}`},
		})

		// Wait for unread indicator then click the bell
		await expect(page.locator('.notifications .unread-indicator')).toBeVisible({
			timeout: 10000,
		})
		await page.locator('.notifications .trigger-button').click()

		// Notification dropdown should contain the team notification
		const notificationsList = page.locator('.notifications .notifications-list')
		await expect(notificationsList).toBeVisible()
		await expect(notificationsList.locator('.single-notification')).toHaveCount(1)
	})

	test('websocket disconnects on logout', async ({authenticatedPage: page}) => {
		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Perform logout — click user menu then logout button
		await page.locator('.navbar .username-dropdown-trigger').click()
		await page.locator('.navbar .dropdown-item').filter({hasText: 'Logout'}).click()

		// After logout, should redirect to login page
		await expect(page).toHaveURL(/\/login/, {timeout: 5000})

		// Verify the notification bell is gone (no authenticated UI)
		await expect(page.locator('.notifications .trigger-button')).toHaveCount(0)
	})
})
