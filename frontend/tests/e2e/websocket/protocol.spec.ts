import {test, expect} from '../../support/fixtures'
import {openWs, waitForMessage, sendMessage, authenticateWs, subscribeWs, collectMessages, closeWs} from '../../support/websocket'
import {UserFactory} from '../../factories/user'
import {TEST_PASSWORD} from '../../support/constants'
import type {APIRequestContext} from '@playwright/test'

/** Login without setting page localStorage — just returns the token. */
async function loginRaw(apiContext: APIRequestContext, user: {username: string}): Promise<{token: string}> {
	const response = await apiContext.post('login', {
		data: {username: user.username, password: TEST_PASSWORD},
	})
	return response.json()
}

test.describe('WebSocket Protocol', () => {

	test.describe('Authentication', () => {
		test('authenticates with valid token', async ({userToken}) => {
			const ws = await openWs()
			try {
				const msg = await authenticateWs(ws, userToken)
				expect(msg.action).toBe('auth.success')
				expect(msg.success).toBe(true)
			} finally {
				closeWs(ws)
			}
		})

		test('rejects invalid token', async () => {
			const ws = await openWs()
			try {
				sendMessage(ws, {action: 'auth', token: 'invalid-token'})
				const msg = await waitForMessage(ws)
				expect(msg.error).toBe('invalid_token')
			} finally {
				closeWs(ws)
			}
		})

		test('closes connection after auth timeout', async () => {
			test.setTimeout(45000)
			const ws = await openWs()
			const closed = new Promise<{code: number; reason: string}>((resolve) => {
				ws.on('close', (code, reason) => {
					resolve({code, reason: reason.toString()})
				})
			})
			const result = await closed
			// websocket StatusPolicyViolation = 1008
			expect(result.code).toBe(1008)
		})

		test('rejects double authentication', async ({userToken}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				sendMessage(ws, {action: 'auth', token: userToken})
				const msg = await waitForMessage(ws)
				expect(msg.error).toBe('already_authenticated')
			} finally {
				closeWs(ws)
			}
		})
	})

	test.describe('Subscribe / Unsubscribe', () => {
		test('subscribes to valid topic', async ({userToken}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				sendMessage(ws, {action: 'subscribe', topic: 'notification.created'})
				// No error response means success — verify by collecting messages
				// for a short window. If there was an error, it would arrive.
				const messages = await collectMessages(ws, 500)
				const errors = messages.filter(m => m.error)
				expect(errors).toHaveLength(0)
			} finally {
				closeWs(ws)
			}
		})

		test('rejects invalid topic', async ({userToken}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				sendMessage(ws, {action: 'subscribe', topic: 'nonexistent.topic'})
				const msg = await waitForMessage(ws)
				expect(msg.error).toBe('invalid_topic')
				expect(msg.topic).toBe('nonexistent.topic')
			} finally {
				closeWs(ws)
			}
		})

		test('requires auth before subscribe', async () => {
			const ws = await openWs()
			try {
				sendMessage(ws, {action: 'subscribe', topic: 'notification.created'})
				const msg = await waitForMessage(ws)
				expect(msg.error).toBe('auth_required')
			} finally {
				closeWs(ws)
			}
		})

		test('unsubscribe stops receiving events', async ({apiContext, userToken, currentUser}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				subscribeWs(ws, 'notification.created')

				// Create a second user to trigger the notification
				const [userA] = await UserFactory.create(1, {id: 100}, false)
				const {token: tokenA} = await loginRaw(apiContext, userA)

				// User A creates a team
				const teamResponse = await apiContext.put('teams', {
					data: {name: 'Test Team'},
					headers: {Authorization: `Bearer ${tokenA}`},
				})
				const team = await teamResponse.json()

				// Unsubscribe before the notification is triggered
				sendMessage(ws, {action: 'unsubscribe', topic: 'notification.created'})
				// Give the server a moment to process the unsubscribe
				await new Promise(r => setTimeout(r, 200))

				// Now add currentUser to team — should NOT receive WS notification
				await apiContext.put(`teams/${team.id}/members`, {
					data: {username: currentUser.username},
					headers: {Authorization: `Bearer ${tokenA}`},
				})

				// Collect messages for 2 seconds — should get none
				const messages = await collectMessages(ws, 2000)
				const notifications = messages.filter(m => m.event === 'notification.created')
				expect(notifications).toHaveLength(0)
			} finally {
				closeWs(ws)
			}
		})
	})

	test.describe('Message Delivery', () => {
		test('receives notification when added to team', async ({apiContext, userToken, currentUser}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				subscribeWs(ws, 'notification.created')

				// Create a second user (the doer)
				const [userA] = await UserFactory.create(1, {id: 100}, false)
				const {token: tokenA} = await loginRaw(apiContext, userA)

				// User A creates a team
				const teamResponse = await apiContext.put('teams', {
					data: {name: 'Notification Test Team'},
					headers: {Authorization: `Bearer ${tokenA}`},
				})
				const team = await teamResponse.json()

				// User A adds currentUser to the team
				const addResponse = await apiContext.put(`teams/${team.id}/members`, {
					data: {username: currentUser.username},
					headers: {Authorization: `Bearer ${tokenA}`},
				})
				expect(addResponse.ok()).toBe(true)

				// currentUser should receive the notification via WebSocket
				const msg = await waitForMessage(ws, 10000)
				expect(msg.event).toBe('notification.created')
				expect(msg.data).toBeDefined()
			} finally {
				closeWs(ws)
			}
		})

		test('doer does not receive own notification', async ({apiContext, userToken}) => {
			const ws = await openWs()
			try {
				await authenticateWs(ws, userToken)
				subscribeWs(ws, 'notification.created')

				// Create a second user
				const [otherUser] = await UserFactory.create(1, {id: 100}, false)

				// currentUser creates a team (they are the doer)
				const teamResponse = await apiContext.put('teams', {
					data: {name: 'Doer Test Team'},
					headers: {Authorization: `Bearer ${userToken}`},
				})
				const team = await teamResponse.json()

				// currentUser adds otherUser — currentUser is the doer
				await apiContext.put(`teams/${team.id}/members`, {
					data: {username: otherUser.username},
					headers: {Authorization: `Bearer ${userToken}`},
				})

				// currentUser should NOT receive a notification (they did the action)
				const messages = await collectMessages(ws, 3000)
				const notifications = messages.filter(m => m.event === 'notification.created')
				expect(notifications).toHaveLength(0)
			} finally {
				closeWs(ws)
			}
		})

		test('multiple connections receive same notification', async ({apiContext, userToken, currentUser}) => {
			const ws1 = await openWs()
			const ws2 = await openWs()
			try {
				// Both connections authenticate as the same user
				await authenticateWs(ws1, userToken)
				await authenticateWs(ws2, userToken)
				subscribeWs(ws1, 'notification.created')
				subscribeWs(ws2, 'notification.created')

				// Create a second user to trigger notification
				const [userA] = await UserFactory.create(1, {id: 100}, false)
				const {token: tokenA} = await loginRaw(apiContext, userA)

				const teamResponse = await apiContext.put('teams', {
					data: {name: 'Multi-Connection Team'},
					headers: {Authorization: `Bearer ${tokenA}`},
				})
				const team = await teamResponse.json()

				await apiContext.put(`teams/${team.id}/members`, {
					data: {username: currentUser.username},
					headers: {Authorization: `Bearer ${tokenA}`},
				})

				// Both connections should receive the notification
				const [msg1, msg2] = await Promise.all([
					waitForMessage(ws1, 10000),
					waitForMessage(ws2, 10000),
				])
				expect(msg1.event).toBe('notification.created')
				expect(msg2.event).toBe('notification.created')
			} finally {
				closeWs(ws1)
				closeWs(ws2)
			}
		})
	})
})
