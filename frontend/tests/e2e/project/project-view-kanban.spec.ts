import {test, expect} from '../../support/fixtures'
import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {createTasksWithPriorities, createTasksWithSearch} from '../../support/filterTestHelpers'
import {updateUserSettings} from '../../support/updateUserSettings'

async function createSingleTaskInBucket(count = 1, attrs = {}) {
	const projects = await ProjectFactory.create(1)
	const views = await ProjectViewFactory.create(1, {
		id: 1,
		project_id: projects[0].id,
		view_kind: 3,
		bucket_configuration_mode: 1,
	})
	const buckets = await BucketFactory.create(2, {
		project_view_id: views[0].id,
	})
	const tasks = await TaskFactory.create(count, {
		project_id: projects[0].id,
		...attrs,
	})
	await TaskBucketFactory.create(1, {
		task_id: tasks[0].id,
		bucket_id: buckets[0].id,
		project_view_id: views[0].id,
	})
	return {
		task: tasks[0],
		view: views[0],
		project: projects[0],
	}
}

async function createTaskWithBuckets(buckets, count = 1) {
	const data = await TaskFactory.create(count, {
		project_id: 1,
	})
	await TaskBucketFactory.truncate()
	for (const t of data) {
		await TaskBucketFactory.create(1, {
			task_id: t.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
	}

	return data
}

test.describe('Project View Kanban', () => {
	let buckets

	test.beforeEach(async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {
			id: 4,
			project_id: projects[0].id,
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		buckets = await BucketFactory.create(2, {
			project_view_id: 4,
		})
	})

	test('Shows all buckets with their tasks', async ({authenticatedPage: page}) => {
		const data = await createTaskWithBuckets(buckets, 10)
		await page.goto('/projects/1/4')

		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[0].title})).toBeVisible()
		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[1].title})).toBeVisible()
		await expect(page.locator('.kanban .bucket').first()).toContainText(data[0].title)
	})

	test('Can add a new task to a bucket', async ({authenticatedPage: page}) => {
		await createTaskWithBuckets(buckets, 2)
		await page.goto('/projects/1/4')

		await page.locator('.kanban .bucket').filter({hasText: buckets[0].title}).locator('.bucket-footer .button').filter({hasText: 'Add another task'}).click()
		await page.locator('.kanban .bucket').filter({hasText: buckets[0].title}).locator('.bucket-footer .field .control input.input').fill('New Task')
		await page.locator('.kanban .bucket').filter({hasText: buckets[0].title}).locator('.bucket-footer .field .control input.input').press('Enter')

		await expect(page.locator('.kanban .bucket').first()).toContainText('New Task')
	})

	test('Can create a new bucket', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		await page.locator('.kanban .bucket.new-bucket .button').click()
		await page.locator('.kanban .bucket.new-bucket input.input').fill('New Bucket')
		await page.locator('.kanban .bucket.new-bucket input.input').press('Enter')

		await expect(page.locator('.kanban .bucket .title').filter({hasText: 'New Bucket'})).toBeVisible()
	})

	test('Can set a bucket limit', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		const bucketDropdown = page.locator('.kanban .bucket .bucket-header .dropdown.options').first()
		await bucketDropdown.locator('.dropdown-trigger').click()
		await bucketDropdown.locator('.dropdown-menu .dropdown-item').filter({hasText: 'Limit: Not Set'}).click()
		await bucketDropdown.locator('.dropdown-menu .field input.input').fill('3')
		await bucketDropdown.locator('.dropdown-menu .field .control .button').click()

		// Wait for the limit to be saved - the dropdown closes and limit is shown
		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.kanban .bucket .bucket-header span.limit').first()).toBeVisible()
		await expect(page.locator('.kanban .bucket .bucket-header span.limit').first()).toContainText('/3')
	})

	test('Can rename a bucket', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		const titleElement = page.locator('.kanban .bucket .bucket-header .title').first()
		await titleElement.click()
		await titleElement.fill('New Bucket Title')
		await titleElement.press('Enter')
		await expect(titleElement).toContainText('New Bucket Title')
	})

	test('Can delete a bucket', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		await page.locator('.kanban .bucket .bucket-header .dropdown.options .dropdown-trigger').first().click()
		await page.locator('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item').filter({hasText: 'Delete'}).click()
		await expect(page.locator('.modal-mask .modal-container .modal-content .modal-header')).toContainText('Delete the bucket')
		await page.locator('.modal-mask .modal-container .modal-content .actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[0].title})).not.toBeVisible()
		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[1].title})).toBeVisible()
	})

	test('Can drag tasks around', async ({authenticatedPage: page}) => {
		const tasks = await createTaskWithBuckets(buckets, 2)
		await page.goto('/projects/1/4')

		const sourceTask = page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title}).first()
		const targetBucket = page.locator('.kanban .bucket:nth-child(2) .tasks')
		await sourceTask.dragTo(targetBucket)

		await expect(page.locator('.kanban .bucket:nth-child(2) .tasks')).toContainText(tasks[0].title)
		await expect(page.locator('.kanban .bucket:nth-child(1) .tasks')).not.toContainText(tasks[0].title)
	})

	test('Should navigate to the task when the task card is clicked', async ({authenticatedPage: page}) => {
		const tasks = await createTaskWithBuckets(buckets, 5)
		await page.goto('/projects/1/4')

		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title}).click()

		await expect(page).toHaveURL(new RegExp(`/tasks/${tasks[0].id}`), {timeout: 1000})
	})

	test('Should remove a task from the kanban board when moving it to another project', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(2)
		const views = await ProjectViewFactory.create(2, {
			project_id: '{increment}',
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		await BucketFactory.create(2)
		const tasks = await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		await TaskBucketFactory.create(5, {
			project_view_id: 1,
		})
		const task = tasks[0]
		await page.goto('/projects/1/' + views[0].id)

		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()

		await page.locator('.task-view .action-buttons .button', {timeout: 3000}).filter({hasText: /^Move$/}).click()
		const multiselectInput = page.locator('.task-view .content.details .field .multiselect.control .input-wrapper input')
		await expect(multiselectInput).toBeVisible({timeout: 5000})
		await multiselectInput.click()
		await multiselectInput.pressSequentially(projects[1].title)
		// Wait for search results to appear before clicking
		const searchResults = page.locator('.task-view .content.details .field .multiselect.control .search-results')
		await searchResults.waitFor({state: 'visible'})
		await searchResults.locator('> *').first().click()

		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 1000})
		await page.goBack()
		const bucketCount = await page.locator('.kanban .bucket').count()
		for (let i = 0; i < bucketCount; i++) {
			await expect(page.locator('.kanban .bucket').nth(i)).not.toContainText(task.title)
		}
	})

	test('Shows a button to filter the kanban board', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		await expect(page.locator('.project-kanban .filter-container .base-button')).toBeVisible()
	})

	test('Should remove a task from the board when deleting it', async ({authenticatedPage: page}) => {
		const {task, view} = await createSingleTaskInBucket(5)
		await page.goto(`/projects/1/${view.id}`)

		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
		await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'})).toBeVisible()
		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'}).click()
		await expect(page.locator('.modal-mask .modal-container .modal-content .modal-header')).toContainText('Delete this task')
		await page.locator('.modal-mask .modal-container .modal-content .actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')

		await page.goBack()
		const bucketCount = await page.locator('.kanban .bucket').count()
		for (let i = 0; i < bucketCount; i++) {
			await expect(page.locator('.kanban .bucket').nth(i)).not.toContainText(task.title)
		}
	})

	test('Should show a task description icon if the task has a description', async ({authenticatedPage: page}) => {
		const {task, view} = await createSingleTaskInBucket(1, {
			description: 'Lorem Ipsum',
		})
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)

		await page.goto(`/projects/${task.project_id}/${view.id}`)
		await loadTasksPromise

		await expect(page.locator('.bucket .tasks .task .footer .icon svg')).toBeVisible()
	})

	test('Should not show a task description icon if the task has an empty description', async ({authenticatedPage: page}) => {
		const {task, view} = await createSingleTaskInBucket(1, {
			description: '',
		})
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)

		await page.goto(`/projects/${task.project_id}/${view.id}`)
		await loadTasksPromise

		await expect(page.locator('.bucket .tasks .task .footer .icon svg')).not.toBeVisible()
	})

	test('Should not show a task description icon if the task has a description containing only an empty p tag', async ({authenticatedPage: page}) => {
		const {task, view} = await createSingleTaskInBucket(1, {
			description: '<p></p>',
		})
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)

		await page.goto(`/projects/${task.project_id}/${view.id}`)
		await loadTasksPromise

		await expect(page.locator('.bucket .tasks .task .footer .icon svg')).not.toBeVisible()
	})

	test('Should respect filter query parameter from URL', async ({authenticatedPage: page}) => {
		// Create buckets first
		const projects = await ProjectFactory.create(1)
		const views = await ProjectViewFactory.create(1, {
			id: 4,
			project_id: 1,
			view_kind: 3,
		})
		const buckets = await BucketFactory.create(2, {
			project_view_id: 4,
		})

		const {highPriorityTasks, lowPriorityTasks} = await createTasksWithPriorities(buckets)

		await page.goto('/projects/1/4?filter=priority%20>=%204')

		await expect(page).toHaveURL(/filter=priority/)

		// Wait for tasks to load and verify high priority tasks are visible
		await expect(page.locator('.kanban')).toContainText(highPriorityTasks[0].title, {timeout: 10000})
		await expect(page.locator('.kanban')).toContainText(highPriorityTasks[1].title)

		// Verify low priority tasks are not visible
		await expect(page.locator('.kanban')).not.toContainText(lowPriorityTasks[0].title)
		await expect(page.locator('.kanban')).not.toContainText(lowPriorityTasks[1].title)
	})

	test('Should respect search query parameter from URL', async ({authenticatedPage: page}) => {
		// Create buckets first
		const projects = await ProjectFactory.create(1)
		const views = await ProjectViewFactory.create(1, {
			id: 4,
			project_id: 1,
			view_kind: 3,
		})
		const buckets = await BucketFactory.create(2, {
			project_view_id: 4,
		})

		const {searchableTask} = await createTasksWithSearch(buckets)

		await page.goto('/projects/1/4?s=meeting')

		await expect(page).toHaveURL(/s=meeting/)

		// Wait for search results to load and verify searchable task is visible
		await expect(page.locator('.kanban')).toContainText(searchableTask.title, {timeout: 10000})

		// Verify only one task is shown (the search result) - count task headings
		await expect(page.locator('main h2')).toHaveCount(1)
	})

	test('Should not show task count by default when bucket has no limit', async ({authenticatedPage: page}) => {
		await createTaskWithBuckets(buckets, 5)
		await page.goto('/projects/1/4')

		// Wait for buckets to load
		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[0].title})).toBeVisible()

		// Verify the task count span is not visible when no limit is set
		await expect(page.locator('.kanban .bucket .bucket-header span.limit').first()).not.toBeVisible()
	})

	test('Should show task count when alwaysShowBucketTaskCount setting is enabled', async ({authenticatedPage: page, apiContext, userToken}) => {
		await createTaskWithBuckets(buckets, 5)

		// Enable the alwaysShowBucketTaskCount setting
		await updateUserSettings(apiContext, userToken, {
			frontendSettings: {
				alwaysShowBucketTaskCount: true,
			},
		})

		await page.goto('/projects/1/4')

		// Wait for buckets to load
		await expect(page.locator('.kanban .bucket .title').filter({hasText: buckets[0].title})).toBeVisible()

		// Verify the task count is shown (without limit, just the count)
		const limitSpan = page.locator('.kanban .bucket .bucket-header span.limit').first()
		await expect(limitSpan).toBeVisible()
		// Should show just the count (5) without a limit
		await expect(limitSpan).toContainText('5')
		// Should not contain a slash (no limit set)
		await expect(limitSpan).not.toContainText('/')
	})

	test('Should show count/limit format when bucket has a limit set', async ({authenticatedPage: page}) => {
		// Create tasks in the bucket
		await createTaskWithBuckets(buckets, 2)
		await page.goto('/projects/1/4')

		// Set a bucket limit
		const bucketDropdown = page.locator('.kanban .bucket .bucket-header .dropdown.options').first()
		await bucketDropdown.locator('.dropdown-trigger').click()
		await bucketDropdown.locator('.dropdown-menu .dropdown-item').filter({hasText: 'Limit: Not Set'}).click()
		await bucketDropdown.locator('.dropdown-menu .field input.input').fill('10')
		await bucketDropdown.locator('.dropdown-menu .field .control .button').click()

		// Wait for the limit to be saved
		await expect(page.locator('.global-notification')).toContainText('Success')

		// Verify the count/limit format is shown (2/10)
		const limitSpan = page.locator('.kanban .bucket .bucket-header span.limit').first()
		await expect(limitSpan).toBeVisible()
		await expect(limitSpan).toContainText('2/10')
	})
})
