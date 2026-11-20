import {test, expect} from '../../support/fixtures'
import {BucketFactory} from '../../factories/bucket'
import {TaskFactory} from '../../factories/task'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {createProjects} from './prepareProjects'

async function openFilterPopup(page) {
	await page.locator('.filter-container button').filter({hasText: 'Filters'}).click()
	await expect(page.locator('.filter-popup')).toBeVisible()
}

async function openAndSetFilters(page, filter = 'done = true') {
	await openFilterPopup(page)
	await page.getByRole('tab', {name: 'Advanced', exact: true}).click()
	await page.locator('.filter-popup .filter-input .ProseMirror').fill(filter)
	await page.locator('.filter-popup button').filter({hasText: 'Show results'}).click()
}

test.describe('Filter Persistence Across Views', () => {
	let seededTasks

	test.beforeEach(async ({authenticatedPage: page}) => {
		await createProjects()
		seededTasks = await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
			title: 'Test Task {increment}',
		})
		await page.goto('/projects/1/1')
	})

	test('should persist filters in List view after page refresh', async ({authenticatedPage: page}) => {
		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should persist filters in Table view after page refresh', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/3')

		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should persist filters in Kanban view after page refresh', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should keep the kanban filter applied while a task modal is open', async ({authenticatedPage: page}) => {
		const buckets = await BucketFactory.create(2, {
			project_view_id: 4,
		})
		await TaskBucketFactory.create(seededTasks.length, {
			task_id: (i: number) => seededTasks[i - 1].id,
			bucket_id: buckets[0].id,
			project_view_id: 4,
		})
		const [doneTask] = await TaskFactory.create(1, {
			id: 6,
			project_id: 1,
			title: 'Done Task',
			done: true,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: doneTask.id,
			bucket_id: buckets[0].id,
			project_view_id: 4,
		}, false)

		await page.goto('/projects/1/4')
		const cards = page.locator('.kanban .bucket .tasks .task')
		await expect(cards).toHaveCount(seededTasks.length + 1)

		await openAndSetFilters(page, 'done = false')
		await expect(cards).toHaveCount(seededTasks.length)

		const filterButton = page.locator('.filter-container button').filter({hasText: 'Filters'})
		await expect(filterButton).toHaveClass(/has-filters/)

		// Delay any board refetch so an unfiltered request would be visible, like on a slow connection.
		await page.route('**/api/v2/projects/1/views/4/buckets/tasks**', async route => {
			await new Promise(resolve => setTimeout(resolve, 1000))
			await route.continue()
		})

		const card = page.locator(`.kanban .bucket .tasks .task[data-task-id="${seededTasks[0].id}"]`)
		await card.locator('.task-id').click()

		await expect(page.locator('dialog[open] .task-view')).toBeVisible()
		await expect(page).toHaveURL(new RegExp(`/tasks/${seededTasks[0].id}$`))

		// The board behind the modal must stay filtered.
		await expect(filterButton).toHaveClass(/has-filters/)
		await expect(cards.filter({hasText: doneTask.title})).toHaveCount(0)

		await page.locator('dialog[open] .modal-container > .close').click()

		await expect(page).toHaveURL(/filter=/)
		await expect(filterButton).toHaveClass(/has-filters/)
		await expect(cards).toHaveCount(seededTasks.length)
	})

	test('should only show a task without a due date when include nulls is ticked', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			title: 'Overdue Task',
			due_date: new Date(Date.now() - 86_400_000).toISOString(),
		})
		await TaskFactory.create(1, {
			id: 2,
			project_id: 1,
			title: 'No Due Date Task',
		}, false)

		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks')).toContainText('No Due Date Task')

		await openAndSetFilters(page, 'due_date < now')

		await expect(page.locator('.tasks')).toContainText('Overdue Task')
		await expect(page.locator('.tasks')).not.toContainText('No Due Date Task')

		const requestWithNulls = page.waitForRequest(request => {
			const url = new URL(request.url())
			return url.pathname.endsWith('/projects/1/views/1/tasks') &&
				url.searchParams.get('filter_include_nulls') === 'true'
		})

		await openFilterPopup(page)
		const includeNulls = page.locator('.filter-popup .fancy-checkbox').filter({hasText: 'Include Tasks which'})
		// The checkbox stretches across the flex column, so its centre misses the label.
		await includeNulls.locator('label').click()
		await expect(includeNulls.locator('input[type=checkbox]')).toBeChecked()
		await page.locator('.filter-popup button').filter({hasText: 'Show results'}).click()
		await requestWithNulls

		await expect(page.locator('.tasks')).toContainText('No Due Date Task')
		await expect(page.locator('.tasks')).toContainText('Overdue Task')
	})

	test('should handle URL sharing with filters', async ({authenticatedPage: page}) => {
		// Visit URL with pre-existing filter parameters
		await page.goto('/projects/1/4?filter=done%3Dtrue&s=Test')

		// Verify URL parameters are preserved
		await expect(page).toHaveURL(/filter=done%3Dtrue/)
		await expect(page).toHaveURL(/s=Test/)

		// Switch views and verify parameters persist
		await page.goto('/projects/1/3?filter=done%3Dtrue&s=Test')
		await expect(page).toHaveURL(/filter=done%3Dtrue/)
		await expect(page).toHaveURL(/s=Test/)
	})
})
