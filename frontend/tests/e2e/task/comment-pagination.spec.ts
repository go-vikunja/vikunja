import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Task comment pagination', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1)
		createDefaultViews(1)
		await TaskFactory.create(1, {id: 1})
		TaskCommentFactory.truncate()
	})

	test('shows pagination when more comments than configured page size', async ({authenticatedPage: page, apiContext}) => {
		const response = await apiContext.get('info')
		const body = await response.json()
		const pageSize = body.max_items_per_page
		await TaskCommentFactory.create(pageSize + 10)
		await page.goto('/tasks/1')
		await expect(page.locator('.task-view .comments nav.pagination')).toBeVisible()
	})

	test('hides pagination when comments equal or fewer than configured page size', async ({authenticatedPage: page, apiContext}) => {
		const response = await apiContext.get('info')
		const body = await response.json()
		const pageSize = body.max_items_per_page
		await TaskCommentFactory.create(Math.max(1, pageSize - 10))
		await page.goto('/tasks/1')
		await expect(page.locator('.task-view .comments nav.pagination')).not.toBeVisible()
	})
})
