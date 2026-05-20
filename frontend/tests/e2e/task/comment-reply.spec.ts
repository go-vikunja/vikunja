import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Reply to a task comment', () => {
	test.beforeEach(async ({authenticatedPage: page, currentUser}) => {
		await ProjectFactory.create(1, {owner_id: currentUser.id})
		await createDefaultViews(1)
		await TaskFactory.create(1, {id: 1, created_by_id: currentUser.id})
	})

	test('Reply action prefills the editor with a quoted blockquote and the saved reply renders an author header + chevron that jumps to the original', async ({authenticatedPage: page, currentUser}) => {
		await TaskCommentFactory.create(1, {
			id: 1,
			task_id: 1,
			author_id: currentUser.id,
			comment: 'Original message that we will quote.',
		})

		await page.goto('/tasks/1')
		await page.waitForLoadState('networkidle')

		const originalComment = page.locator('#comment-1')
		await expect(originalComment).toBeVisible({timeout: 10000})

		// The Reply action lives in the per-comment bottom-actions list.
		await originalComment.getByRole('button', {name: 'Reply', exact: true}).click()

		// The new-comment editor (the contenteditable one) should now contain
		// the prefilled blockquote pointing back at comment 1.
		const newCommentEditor = page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror[contenteditable="true"]').last()
		await expect(newCommentEditor).toBeVisible()
		await expect(newCommentEditor.locator('blockquote[data-comment-id="1"]')).toBeVisible()
		await expect(newCommentEditor.locator('blockquote[data-comment-id="1"]')).toContainText('Original message that we will quote.')

		// Append a reply body after the auto-inserted paragraph.
		await newCommentEditor.click()
		await page.keyboard.press('End')
		await page.keyboard.type('Thanks for that!')

		await page.getByRole('button', {name: 'Comment', exact: true}).click()

		// The newly-rendered reply should carry the quote header + chevron.
		const reply = page.locator('.task-view .comments .media.comment').nth(1)
		await expect(reply).toBeVisible()
		const quote = reply.locator('blockquote.comment-quote[data-comment-id="1"]')
		await expect(quote).toBeVisible()
		await expect(quote.locator('.comment-quote__jump')).toBeVisible()

		// Clicking the chevron scrolls to and briefly highlights the original.
		await quote.locator('.comment-quote__jump').click()
		await expect(originalComment).toHaveClass(/comment-highlight/, {timeout: 2000})
	})
})
