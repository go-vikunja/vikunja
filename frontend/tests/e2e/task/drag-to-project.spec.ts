import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {SavedFilterFactory} from '../../factories/saved_filter'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'

async function createProjectsWithTasks() {
	// Create two projects
	const projects = await ProjectFactory.create(2, {
		title: i => i === 0 ? 'Source Project' : 'Target Project',
	})

	// Create views for both projects
	await ProjectViewFactory.truncate()

	// List view for source project
	const sourceListView = await ProjectViewFactory.create(1, {
		id: 1,
		project_id: projects[0].id,
		view_kind: 0,
	}, false)

	// Kanban view for source project
	const sourceKanbanView = await ProjectViewFactory.create(1, {
		id: 2,
		project_id: projects[0].id,
		view_kind: 3,
		bucket_configuration_mode: 1,
	}, false)

	// List view for target project
	await ProjectViewFactory.create(1, {
		id: 3,
		project_id: projects[1].id,
		view_kind: 0,
	}, false)

	// Create bucket for kanban view
	const buckets = await BucketFactory.create(1, {
		project_view_id: 2,
	})

	// Create tasks in source project
	await TaskFactory.truncate()
	const tasks = await TaskFactory.create(3, {
		id: '{increment}',
		title: i => `Task ${i + 1}`,
		project_id: projects[0].id,
	})

	// Assign tasks to bucket for kanban view
	await TaskBucketFactory.truncate()
	for (const task of tasks) {
		await TaskBucketFactory.create(1, {
			task_id: task.id,
			bucket_id: buckets[0].id,
			project_view_id: 2,
		}, false)
	}

	return {
		sourceProject: projects[0],
		targetProject: projects[1],
		sourceListView: sourceListView[0],
		sourceKanbanView: sourceKanbanView[0],
		tasks,
		bucket: buckets[0],
	}
}

test.describe('Drag Task to Project in Sidebar', () => {
	test.describe('From List View', () => {
		test('Can drag a task to another project in the sidebar', async ({authenticatedPage: page}) => {
			const {sourceProject, targetProject, sourceListView, tasks} = await createProjectsWithTasks()

			await page.goto(`/projects/${sourceProject.id}/${sourceListView.id}`)

			// Wait for tasks to load
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// Find the task and the target project in sidebar (use li selector to avoid matching task elements)
			const task = page.locator('.tasks .single-task').filter({hasText: tasks[0].title})
			const targetProjectInSidebar = page.locator('li[data-project-id="' + targetProject.id + '"]')

			// Drag task to target project
			await task.dragTo(targetProjectInSidebar)

			// Verify success notification
			await expect(page.locator('.global-notification')).toContainText('moved to')

			// Verify task is removed from the list
			await expect(page.locator('.tasks')).not.toContainText(tasks[0].title)

			// Verify task appears in target project
			await page.goto(`/projects/${targetProject.id}/3`)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)
		})

		test('Does not move task when dropped on the same project', async ({authenticatedPage: page}) => {
			const {sourceProject, sourceListView, tasks} = await createProjectsWithTasks()

			await page.goto(`/projects/${sourceProject.id}/${sourceListView.id}`)

			// Wait for tasks to load
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// Find the task and the source project in sidebar (use li selector to avoid matching task elements)
			const task = page.locator('.tasks .single-task').filter({hasText: tasks[0].title})
			const sourceProjectInSidebar = page.locator('li[data-project-id="' + sourceProject.id + '"]')

			// Drag task to the same project
			await task.dragTo(sourceProjectInSidebar)

			// Task should still be in the list (no move occurred)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// No success notification should appear for same-project drop
			await expect(page.locator('.global-notification')).not.toContainText('moved to')
		})
	})

	test.describe('From Kanban View', () => {
		test('Can drag a task to another project in the sidebar', async ({authenticatedPage: page}) => {
			const {sourceProject, targetProject, sourceKanbanView, tasks} = await createProjectsWithTasks()

			await page.goto(`/projects/${sourceProject.id}/${sourceKanbanView.id}`)

			// Wait for kanban to load
			await expect(page.locator('.kanban .bucket .tasks')).toContainText(tasks[0].title)

			// Find the task and the target project in sidebar (use li selector to avoid matching task elements)
			const task = page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title})
			const targetProjectInSidebar = page.locator('li[data-project-id="' + targetProject.id + '"]')

			// Drag task to target project
			await task.dragTo(targetProjectInSidebar)

			// Verify success notification
			await expect(page.locator('.global-notification')).toContainText('moved to')

			// Verify task is removed from the kanban board
			await expect(page.locator('.kanban .bucket .tasks')).not.toContainText(tasks[0].title)

			// Verify task appears in target project
			await page.goto(`/projects/${targetProject.id}/3`)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)
		})

		test('Does not move task when dropped on the same project', async ({authenticatedPage: page}) => {
			const {sourceProject, sourceKanbanView, tasks} = await createProjectsWithTasks()

			await page.goto(`/projects/${sourceProject.id}/${sourceKanbanView.id}`)

			// Wait for kanban to load
			await expect(page.locator('.kanban .bucket .tasks')).toContainText(tasks[0].title)

			// Find the task and the source project in sidebar (use li selector to avoid matching task elements)
			const task = page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title})
			const sourceProjectInSidebar = page.locator('li[data-project-id="' + sourceProject.id + '"]')

			// Drag task to the same project
			await task.dragTo(sourceProjectInSidebar)

			// Task should still be in the kanban (no move occurred)
			await expect(page.locator('.kanban .bucket .tasks')).toContainText(tasks[0].title)

			// No success notification should appear for same-project drop
			await expect(page.locator('.global-notification')).not.toContainText('moved to')
		})
	})

	test.describe('Invalid Drop Targets', () => {
		test('Does not move task when dropped on a saved filter', async ({authenticatedPage: page}) => {
			// Create source project with tasks
			const projects = await ProjectFactory.create(1, {
				title: 'Source Project',
			})

			await ProjectViewFactory.truncate()
			const sourceListView = await ProjectViewFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				view_kind: 0,
			}, false)

			// Create a saved filter (shows as pseudo-project with negative ID in sidebar)
			await SavedFilterFactory.create(1, {
				id: 1,
				title: 'My Saved Filter',
				owner_id: 1,
			})

			await TaskFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				title: 'Test Task',
				project_id: projects[0].id,
			})

			await page.goto(`/projects/${projects[0].id}/${sourceListView[0].id}`)

			// Wait for tasks to load
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// Find the task and the saved filter in sidebar
			// Saved filters have negative IDs: id * -1 - 1, so filter id=1 -> project id=-2
			const task = page.locator('.tasks .single-task').filter({hasText: tasks[0].title})
			const savedFilterInSidebar = page.locator('li[data-project-id="-2"]')

			// Verify the saved filter is visible in sidebar
			await expect(savedFilterInSidebar).toBeVisible()

			// Drag task to the saved filter
			await task.dragTo(savedFilterInSidebar)

			// Task should still be in the list (saved filters cannot accept tasks)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// No success notification should appear
			await expect(page.locator('.global-notification')).not.toContainText('moved to')
		})

		test('Does not move task when dropped on a read-only shared project', async ({authenticatedPage: page}) => {
			// Create a second user who will own the read-only project
			await UserFactory.create(2)

			// Create source project (owned by user 1)
			const sourceProject = await ProjectFactory.create(1, {
				id: 1,
				title: 'Source Project',
				owner_id: 1,
			})

			// Create target project (owned by user 2, shared read-only to user 1)
			const readOnlyProject = await ProjectFactory.create(1, {
				id: 2,
				title: 'Read Only Project',
				owner_id: 2,
			}, false)

			// Share the project read-only to user 1 (permission 0 = read)
			await UserProjectFactory.create(1, {
				project_id: 2,
				user_id: 1,
				permission: 0,
			})

			await ProjectViewFactory.truncate()
			const sourceListView = await ProjectViewFactory.create(1, {
				id: 1,
				project_id: sourceProject[0].id,
				view_kind: 0,
			}, false)

			// Create a view for the read-only project so it shows in sidebar
			await ProjectViewFactory.create(1, {
				id: 2,
				project_id: readOnlyProject[0].id,
				view_kind: 0,
			}, false)

			await TaskFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				title: 'Test Task',
				project_id: sourceProject[0].id,
			})

			await page.goto(`/projects/${sourceProject[0].id}/${sourceListView[0].id}`)

			// Wait for tasks to load
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// Find the task and the read-only project in sidebar
			const task = page.locator('.tasks .single-task').filter({hasText: tasks[0].title})
			const readOnlyProjectInSidebar = page.locator('li[data-project-id="' + readOnlyProject[0].id + '"]')

			// Verify the read-only project is visible in sidebar
			await expect(readOnlyProjectInSidebar).toBeVisible()

			// Drag task to the read-only project
			await task.dragTo(readOnlyProjectInSidebar)

			// Task should still be in the list (read-only projects cannot accept tasks)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// No success notification should appear (visual highlight doesn't show for read-only)
			await expect(page.locator('.global-notification')).not.toContainText('moved to')
		})

		test('Shows error notification when task move fails', async ({authenticatedPage: page}) => {
			const {sourceProject, targetProject, sourceListView, tasks} = await createProjectsWithTasks()

			await page.goto(`/projects/${sourceProject.id}/${sourceListView.id}`)

			// Wait for tasks to load
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)

			// Intercept the task update API call and return an error
			await page.route('**/api/v1/tasks/*', async (route) => {
				if (route.request().method() === 'POST') {
					await route.fulfill({
						status: 500,
						contentType: 'application/json',
						body: JSON.stringify({
							code: 500,
							message: 'Internal server error',
						}),
					})
				} else {
					await route.continue()
				}
			})

			// Find the task and the target project in sidebar
			const task = page.locator('.tasks .single-task').filter({hasText: tasks[0].title})
			const targetProjectInSidebar = page.locator('li[data-project-id="' + targetProject.id + '"]')

			// Drag task to target project
			await task.dragTo(targetProjectInSidebar)

			// Verify error notification appears
			await expect(page.locator('.global-notification .vue-notification.error')).toBeVisible()

			// Task should still be in the list (move failed)
			await expect(page.locator('.tasks')).toContainText(tasks[0].title)
		})
	})
})
