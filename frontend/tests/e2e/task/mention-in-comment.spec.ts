import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Mention in task comment', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1)
		await createDefaultViews(1)
		await TaskFactory.create(1, {id: 1})
	})

	test('typing @ in comment editor does not throw TypeError', async ({authenticatedPage: page}) => {
		// Collect console errors
		const consoleErrors: string[] = []
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text())
			}
		})

		await page.goto('/tasks/1')
		await page.waitForLoadState('networkidle')

		// Wait for comment input editor to be visible (the editable one)
		const commentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]')
		await expect(commentEditor).toBeVisible({timeout: 10000})

		// Click to focus the editor
		await commentEditor.click()

		// Type @ to trigger mention suggestion
		await commentEditor.pressSequentially('@', {delay: 50})

		// Wait a bit for any async operations
		await page.waitForTimeout(500)

		// Type more characters to trigger updates
		await commentEditor.pressSequentially('test', {delay: 50})

		// Wait for debounce and any potential errors
		await page.waitForTimeout(500)

		// Press Escape to close any mention popup
		await page.keyboard.press('Escape')

		// Verify no TypeErrors related to mention component were logged
		const mentionErrors = consoleErrors.filter(err =>
			err.includes('TypeError') &&
			(err.includes('updateProps') || err.includes('onKeyDown') || err.includes('ref')),
		)

		expect(mentionErrors).toHaveLength(0)
	})

	test('suggests a project member and inserts the mention', async ({authenticatedPage: page, currentUser}) => {
		const [member] = await UserFactory.create(1, {id: currentUser.id + 100}, false)
		await UserProjectFactory.create(1, {project_id: 1, user_id: member.id})

		await page.goto('/tasks/1')

		const commentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]')
		await expect(commentEditor).toBeVisible({timeout: 10000})
		await commentEditor.click()
		await commentEditor.pressSequentially(`@${member.username.substring(0, 8)}`, {delay: 50})

		const suggestion = page.locator('.mention-items .mention-item').filter({hasText: member.username})
		await expect(suggestion).toBeVisible({timeout: 5000})
		await suggestion.click()

		await expect(commentEditor.locator('.mention-user')).toContainText(member.username)
	})

	test('can type mention without error notifications appearing', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await page.waitForLoadState('networkidle')

		// Wait for comment input editor to be visible (the editable one)
		const commentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]')
		await expect(commentEditor).toBeVisible({timeout: 10000})

		// Click to focus the editor
		await commentEditor.click()

		// Type @ to trigger mention suggestion
		await commentEditor.pressSequentially('@user', {delay: 50})

		// Wait for potential error notifications to appear
		await page.waitForTimeout(1000)

		// Verify no error notification appeared
		const errorNotification = page.locator('.global-notification.is-danger, .global-notification.error')
		await expect(errorNotification).not.toBeVisible()
	})
})
