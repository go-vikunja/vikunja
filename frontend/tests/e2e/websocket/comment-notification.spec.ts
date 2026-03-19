import {test, expect} from '../../support/fixtures'
import {openWs, waitForMessage, authenticateWs, subscribeWs, closeWs} from '../../support/websocket'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskFactory} from '../../factories/task'
import {UserProjectFactory} from '../../factories/users_project'
import {TEST_PASSWORD} from '../../support/constants'
import type {APIRequestContext} from '@playwright/test'

async function loginRaw(apiContext: APIRequestContext, user: {username: string}): Promise<{token: string}> {
	const response = await apiContext.post('login', {
		data: {username: user.username, password: TEST_PASSWORD},
	})
	return response.json()
}

test.describe('WebSocket Comment Notifications', () => {

	test('receives notification when mentioned in a task comment', async ({apiContext, userToken, currentUser}) => {
		const ws = await openWs()
		try {
			await authenticateWs(ws, userToken)
			subscribeWs(ws, 'notification.created')

			// Create a second user who will post the comment
			const [commenter] = await UserFactory.create(1, {id: 100}, false)
			const {token: commenterToken} = await loginRaw(apiContext, commenter)

			// Seed a project owned by the commenter with a task
			await ProjectFactory.create(1, {id: 100, owner_id: 100}, false)
			await ProjectViewFactory.create(1, {id: 100, project_id: 100}, false)
			await TaskFactory.create(1, {id: 100, project_id: 100, created_by_id: 100}, false)

			// Share the project with currentUser so the mention access check passes
			await UserProjectFactory.create(1, {id: 100, project_id: 100, user_id: 1}, false)

			// Commenter posts a comment mentioning currentUser
			const commentBody = `<p>Hey <mention-user data-id="${currentUser.username}">@${currentUser.username}</mention-user> check this out</p>`
			const commentResponse = await apiContext.put('tasks/100/comments', {
				data: {comment: commentBody},
				headers: {Authorization: `Bearer ${commenterToken}`},
			})
			expect(commentResponse.ok()).toBe(true)

			// currentUser should receive the notification via WebSocket
			const msg = await waitForMessage(ws, 15000)
			expect(msg.event).toBe('notification.created')
			expect(msg.data).toBeDefined()
		} finally {
			closeWs(ws)
		}
	})
})
