import {test, expect} from '../../support/fixtures'
import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {TaskRelationFactory} from '../../factories/task_relation'

async function createKanbanTaskInBucket() {
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
	const tasks = await TaskFactory.create(1, {
		project_id: projects[0].id,
	})
	await TaskBucketFactory.create(1, {
		task_id: tasks[0].id,
		bucket_id: buckets[0].id,
		project_view_id: views[0].id,
	})
	return {
		project: projects[0],
		view: views[0],
		buckets,
		task: tasks[0],
	}
}

test.describe('Task Bucket Select', () => {
	test('Shows the current bucket name when opening a task from a kanban view', async ({authenticatedPage: page}) => {
		const {project, view, buckets, task} = await createKanbanTaskInBucket()

		await page.goto(`/projects/${project.id}/${view.id}`)
		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
		await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

		await expect(page.locator('.task-view .subtitle')).toContainText(buckets[0].title)
	})

	test('Can change the bucket from the task detail view', async ({authenticatedPage: page}) => {
		const {project, view, buckets, task} = await createKanbanTaskInBucket()

		await page.goto(`/projects/${project.id}/${view.id}`)
		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
		await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

		// Click the bucket name to open the dropdown
		await page.locator('.task-view .subtitle .bucket-name').click()
		// Select the other bucket
		await page.locator('.task-view .subtitle .dropdown-item').filter({hasText: buckets[1].title}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.task-view .subtitle')).toContainText(buckets[1].title)
	})

	test('Does not show the bucket selector when project has no kanban view', async ({authenticatedPage: page}) => {
		// Truncate leftover data from previous tests
		await BucketFactory.truncate()
		await TaskBucketFactory.truncate()
		await TaskRelationFactory.truncate()

		const projects = await ProjectFactory.create(1)
		// Only create a list view, no kanban view
		const views = await ProjectViewFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
			view_kind: 0,
		})
		const tasks = await TaskFactory.create(1, {
			project_id: projects[0].id,
		})

		await page.goto(`/projects/${projects[0].id}/${views[0].id}`)
		await page.locator('.tasks .task').filter({hasText: tasks[0].title}).click()
		await expect(page).toHaveURL(new RegExp(`/tasks/${tasks[0].id}`))

		await expect(page.locator('.task-view .subtitle .bucket-name')).not.toBeVisible()
	})

	test.describe('Multiple kanban views', () => {
		async function createTaskWithMultipleKanbanViews() {
			// Truncate leftover task relations from previous tests
			await TaskRelationFactory.truncate()

			const projects = await ProjectFactory.create(1)
			const listView = (await ProjectViewFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				view_kind: 0,
			}))[0]
			const kanbanView1 = (await ProjectViewFactory.create(1, {
				id: 2,
				project_id: projects[0].id,
				view_kind: 3,
				bucket_configuration_mode: 1,
			}, false))[0]
			const kanbanView2 = (await ProjectViewFactory.create(1, {
				id: 3,
				project_id: projects[0].id,
				view_kind: 3,
				bucket_configuration_mode: 1,
			}, false))[0]
			const bucketsView1 = await BucketFactory.create(2, {
				project_view_id: kanbanView1.id,
			})
			const bucketsView2 = await BucketFactory.create(2, {
				id: (i: number) => i + 2,
				project_view_id: kanbanView2.id,
			}, false)
			const tasks = await TaskFactory.create(1, {
				project_id: projects[0].id,
			})
			await TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: bucketsView1[0].id,
				project_view_id: kanbanView1.id,
			})
			await TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: bucketsView2[0].id,
				project_view_id: kanbanView2.id,
			}, false)
			return {
				project: projects[0],
				listView,
				kanbanView1,
				kanbanView2,
				bucketsView1,
				bucketsView2,
				task: tasks[0],
			}
		}

		test('Does not show the bucket selector when opening a task from the list view', async ({authenticatedPage: page}) => {
			const {project, listView, task} = await createTaskWithMultipleKanbanViews()

			await page.goto(`/projects/${project.id}/${listView.id}`)
			await page.locator('.tasks .task').filter({hasText: task.title}).click()
			await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

			await expect(page.locator('.task-view .subtitle .bucket-name')).not.toBeVisible()
		})

		test('Shows the correct buckets when opening a task from the first kanban view', async ({authenticatedPage: page}) => {
			const {project, kanbanView1, bucketsView1, task} = await createTaskWithMultipleKanbanViews()

			await page.goto(`/projects/${project.id}/${kanbanView1.id}`)
			await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
			await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
			await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

			await expect(page.locator('.task-view .subtitle')).toContainText(bucketsView1[0].title)
			await page.locator('.task-view .subtitle .bucket-name').click()
			await expect(page.locator('.task-view .subtitle .dropdown-item')).toHaveCount(bucketsView1.length)
			for (const bucket of bucketsView1) {
				await expect(page.locator('.task-view .subtitle .dropdown-item').filter({hasText: bucket.title})).toBeVisible()
			}
		})

		test('Shows the correct buckets when opening a task from the second kanban view', async ({authenticatedPage: page}) => {
			const {project, kanbanView2, bucketsView2, task} = await createTaskWithMultipleKanbanViews()

			await page.goto(`/projects/${project.id}/${kanbanView2.id}`)
			await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
			await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
			await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

			await expect(page.locator('.task-view .subtitle')).toContainText(bucketsView2[0].title)
			await page.locator('.task-view .subtitle .bucket-name').click()
			await expect(page.locator('.task-view .subtitle .dropdown-item')).toHaveCount(bucketsView2.length)
			for (const bucket of bucketsView2) {
				await expect(page.locator('.task-view .subtitle .dropdown-item').filter({hasText: bucket.title})).toBeVisible()
			}
		})
	})

	test('Keeps action buttons visible after changing the bucket', async ({authenticatedPage: page}) => {
		const {project, view, buckets, task} = await createKanbanTaskInBucket()

		await page.goto(`/projects/${project.id}/${view.id}`)
		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title})).toBeVisible()
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).click()
		await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))

		// Change the bucket
		await page.locator('.task-view .subtitle .bucket-name').click()
		await page.locator('.task-view .subtitle .dropdown-item').filter({hasText: buckets[1].title}).click()
		await expect(page.locator('.global-notification')).toContainText('Success')

		// Action buttons should still be visible
		await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Done'})).toBeVisible()
	})
})
