import {test, expect} from '../../support/fixtures'
import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskBucketFactory} from '../../factories/task_buckets'

test.describe('Bucket Select in Task Detail', () => {
	test('Can change bucket from task detail view', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		const views = await ProjectViewFactory.create(1, {
			id: 4,
			project_id: projects[0].id,
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		const buckets = await BucketFactory.create(2, {
			project_view_id: views[0].id,
		})
		const tasks = await TaskFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
		})
		await TaskBucketFactory.create(1, {
			task_id: tasks[0].id,
			bucket_id: buckets[0].id,
			project_view_id: views[0].id,
		})

		// Open task detail view
		await page.goto(`/tasks/${tasks[0].id}`)

		// Wait for the task to load
		await expect(page.locator('.task-view h1.title.input')).toContainText(tasks[0].title)

		// Click "Set Bucket" button
		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Bucket'}).click()

		// The bucket selector should appear
		const bucketColumn = page.locator('.task-view .columns.details .column').filter({hasText: 'Bucket'})
		await expect(bucketColumn).toBeVisible()

		const bucketSelect = bucketColumn.locator('[data-cy="bucket-select"]')
		await expect(bucketSelect).toBeVisible()

		// Screenshot: bucket selector visible in task detail
		await page.screenshot({path: '../screenshots/bucket-selector-visible.png'})

		// Change to the second bucket
		await bucketSelect.selectOption({label: buckets[1].title})

		// Should show success notification
		await expect(page.locator('.global-notification')).toContainText('Success')

		// Verify that the task detail sidebar still shows all UI elements after bucket change
		// (regression test for bug where Object.assign overwrote relations with empty values)
		const actionButtons = page.locator('.task-view .action-buttons')
		await expect(actionButtons).toBeVisible()
		// The action buttons section should still contain multiple buttons (not be empty)
		const buttonCount = await actionButtons.locator('.button').count()
		expect(buttonCount).toBeGreaterThan(2)

		// Screenshot: sidebar still intact after bucket change (cropped to action buttons area)
		await actionButtons.screenshot({path: '../screenshots/sidebar-intact-after-bucket-change.png'})

		// Screenshot: bucket changed with success notification
		await page.screenshot({path: '../screenshots/bucket-changed-success.png'})
	})

	test('Bucket selector is hidden for projects without kanban views', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		// Create only a list view, no kanban view
		await ProjectViewFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
			view_kind: 0,
		})
		const tasks = await TaskFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
		})

		// Open task detail view
		await page.goto(`/tasks/${tasks[0].id}`)

		// Wait for the task to load
		await expect(page.locator('.task-view h1.title.input')).toContainText(tasks[0].title)

		// The "Set Bucket" button should NOT be visible
		await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Bucket'})).not.toBeVisible()

		// Screenshot: no bucket selector for non-kanban project
		await page.screenshot({path: '../screenshots/bucket-selector-hidden-no-kanban.png'})
	})

	test('Changing bucket reflects on kanban board', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		const views = await ProjectViewFactory.create(1, {
			id: 4,
			project_id: projects[0].id,
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		const buckets = await BucketFactory.create(2, {
			project_view_id: views[0].id,
		})
		const tasks = await TaskFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
		})
		await TaskBucketFactory.create(1, {
			task_id: tasks[0].id,
			bucket_id: buckets[0].id,
			project_view_id: views[0].id,
		})

		// First go to the kanban board and verify task is in first bucket
		await page.goto(`/projects/${projects[0].id}/${views[0].id}`)
		await expect(page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title})).toBeVisible()

		// Open the task detail by clicking on it
		await page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title}).click()

		// Wait for task detail to load
		await expect(page.locator('.task-view h1.title.input')).toContainText(tasks[0].title)

		// Click "Set Bucket" and change bucket
		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Bucket'}).click()

		const bucketColumn = page.locator('.task-view .columns.details .column').filter({hasText: 'Bucket'})
		const bucketSelect = bucketColumn.locator('[data-cy="bucket-select"]')
		await expect(bucketSelect).toBeVisible()
		await bucketSelect.selectOption({label: buckets[1].title})

		// Should show success notification
		await expect(page.locator('.global-notification')).toContainText('Success')

		// Close the task detail modal to go back to kanban board
		await page.locator('.modal-container > .close').click()

		// The task should now be in the second bucket
		const secondBucket = page.locator('.kanban .bucket').filter({hasText: buckets[1].title})
		await expect(secondBucket.locator('.tasks .task').filter({hasText: tasks[0].title})).toBeVisible({timeout: 10000})

		// Screenshot: kanban board reflecting the bucket change
		await page.screenshot({path: '../screenshots/kanban-board-bucket-changed.png'})
	})
})
