import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {UserFactory} from '../../factories/user'
import {createDefaultViews} from '../project/prepareProjects'
import {login} from '../../support/authenticateUser'

/**
 * Creates comments with distinct, ordered timestamps so sort order is deterministic.
 * Returns the created comment data array.
 */
async function createCommentsWithTimestamps(count: number, taskId = 1) {
	const comments = []
	const baseDate = new Date('2024-01-01T00:00:00Z')
	for (let i = 1; i <= count; i++) {
		const date = new Date(baseDate.getTime() + i * 60_000) // 1 minute apart
		comments.push({
			id: i,
			comment: `Comment ${i}`,
			author_id: 1,
			task_id: taskId,
			created: date.toISOString(),
			updated: date.toISOString(),
		})
	}
	await TaskCommentFactory.seed(TaskCommentFactory.table, comments)
	return comments
}

function commentLocators(page: Page) {
	return page.locator('.task-view .comments .media.comment[id^="comment-"] .tiptap__editor')
}

test.describe('Comment sort order', () => {
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	test.beforeEach(async ({authenticatedPage}) => {
		await ProjectFactory.create(1)
		await createDefaultViews(1)
		await TaskFactory.create(1, {id: 1})
		await TaskCommentFactory.truncate()
	})

	test('defaults to oldest first', async ({authenticatedPage: page}) => {
		await createCommentsWithTimestamps(3)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		// Wait for comments to load
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Should show oldest (Comment 1) first
		await expect(comments.first()).toContainText('Comment 1')
		await expect(comments.last()).toContainText('Comment 3')

		// Sort button should show "Oldest first" (the current state)
		await expect(page.locator('.comment-sort-button')).toContainText('Oldest first')
	})

	test('toggles to newest first', async ({authenticatedPage: page}) => {
		await createCommentsWithTimestamps(3)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Click the sort button to switch to newest first
		await page.locator('.comment-sort-button').click()

		// Wait for the comments to reload with new order
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')

		// Should now show newest (Comment 3) first
		await expect(comments.first()).toContainText('Comment 3')
		await expect(comments.last()).toContainText('Comment 1')
	})

	test('new comment appears at the top when newest first', async ({authenticatedPage: page}) => {
		await createCommentsWithTimestamps(3)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Switch to newest first
		await page.locator('.comment-sort-button').click()
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')
		await expect(comments.first()).toContainText('Comment 3')

		// Add a new comment
		const newCommentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]')
		await expect(newCommentEditor).toBeVisible({timeout: 10000})
		await newCommentEditor.click()
		await newCommentEditor.fill('Brand new comment')
		await page.locator('.task-view .comments .media.comment .button:not([disabled])').filter({hasText: 'Comment'}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')

		// The new comment should now be at the top (first in the list)
		await expect(comments.first()).toContainText('Brand new comment')
	})

	test('scrolls to top when adding a comment in newest first mode', async ({authenticatedPage: page}) => {
		// Create enough comments to make the page scrollable
		await createCommentsWithTimestamps(10)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Switch to newest first
		await page.locator('.comment-sort-button').click()
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')

		// Add a new comment (the editor is at the bottom)
		const newCommentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]')
		await expect(newCommentEditor).toBeVisible({timeout: 10000})
		await newCommentEditor.click()
		await newCommentEditor.fill('Scroll test comment')
		await page.locator('.task-view .comments .media.comment .button:not([disabled])').filter({hasText: 'Comment'}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')

		// The comments container should be scrolled into view (near the top of viewport)
		const commentsContainer = page.locator('.comments-container')
		await expect(commentsContainer).toBeInViewport({timeout: 5000})
	})

	test('pagination works with sort order', async ({authenticatedPage: page, apiContext}) => {
		const response = await apiContext.get('info')
		const body = await response.json()
		const pageSize = body.max_items_per_page

		await createCommentsWithTimestamps(pageSize + 5)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Default (oldest first): first page should have Comment 1
		await expect(comments.first()).toContainText('Comment 1')
		await expect(page.locator('.task-view .comments nav.pagination')).toBeVisible()

		// Switch to newest first
		await page.locator('.comment-sort-button').click()
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')

		// First page should now have the newest comment
		await expect(comments.first()).toContainText(`Comment ${pageSize + 5}`)
		// Pagination should still be visible
		await expect(page.locator('.task-view .comments nav.pagination')).toBeVisible()
	})

	test('works with initial load (fewer comments than page size)', async ({authenticatedPage: page}) => {
		await createCommentsWithTimestamps(3)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Default: oldest first with initial comments
		await expect(comments.first()).toContainText('Comment 1')
		await expect(comments.last()).toContainText('Comment 3')

		// Toggle to newest first — should reload from API
		await page.locator('.comment-sort-button').click()
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')

		await expect(comments.first()).toContainText('Comment 3')
		await expect(comments.last()).toContainText('Comment 1')
	})

	test('persists sort order setting', async ({authenticatedPage: page}) => {
		await createCommentsWithTimestamps(3)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Switch to newest first
		await page.locator('.comment-sort-button').click()
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')
		await expect(comments.first()).toContainText('Comment 3')

		// Reload the page
		await page.reload()
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Sort order should be preserved
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')
		await expect(comments.first()).toContainText('Comment 3')
		await expect(comments.last()).toContainText('Comment 1')
	})

	test('uses saved setting on page load', async ({page, apiContext}) => {
		// Create a user with commentSortOrder already set to desc
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({commentSortOrder: 'desc'}),
		}))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await TaskFactory.truncate()
		await TaskFactory.create(1, {id: 1, project_id: project.id, created_by_id: user.id})
		await createCommentsWithTimestamps(3)

		await login(page, apiContext, user)
		await page.goto('/tasks/1')

		const comments = commentLocators(page)
		await expect(comments.first()).toBeVisible({timeout: 10000})

		// Should load with newest first based on saved setting
		await expect(page.locator('.comment-sort-button')).toContainText('Newest first')
		await expect(comments.first()).toContainText('Comment 3')
		await expect(comments.last()).toContainText('Comment 1')
	})
})
