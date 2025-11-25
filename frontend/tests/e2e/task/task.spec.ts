import {test, expect} from '../../support/fixtures'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime.js'

dayjs.extend(relativeTime)

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {TaskCommentFactory} from '../../factories/task_comment'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {TaskAssigneeFactory} from '../../factories/task_assignee'
import {LabelFactory} from '../../factories/labels'
import {LabelTaskFactory} from '../../factories/label_task'
import {BucketFactory} from '../../factories/bucket'
import {TaskAttachmentFactory} from '../../factories/task_attachments'
import {TaskReminderFactory} from '../../factories/task_reminders'
import {createDefaultViews} from '../project/prepareProjects'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {pasteFile} from '../../support/commands'
import type {Page} from '@playwright/test'
import {readFileSync} from 'fs'
import {join, dirname} from 'path'
import {fileURLToPath} from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// Type definitions to fix linting errors
interface Project {
	id: number;
	title: string;
	identifier?: string;
}

interface Task {
	id: number;
	title: string;
	description: string;
	project_id: number;
	index: number;
}

interface User {
	id: number;
	username: string;
}

interface Label {
	id: number;
	title: string;
}

interface Bucket {
	id: number;
}

async function addLabelToTaskAndVerify(page: Page, labelTitle: string) {
	await page.locator('.task-view .action-buttons .button').filter({hasText: 'Add Labels'}).click()
	await page.locator('.task-view .details.labels-list .multiselect input').fill(labelTitle)
	await page.locator('.task-view .details.labels-list .multiselect .search-results').locator('> *').first().click()

	await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 4000})
	await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper span.tag')).toBeVisible()
	await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper span.tag')).toContainText(labelTitle)
}

async function uploadAttachmentAndVerify(page: Page, taskId: number) {
	const uploadAttachmentPromise = page.waitForResponse(response =>
		response.url().includes(`/tasks/${taskId}/attachments`) && response.request().method() === 'PUT',
	)
	await page.locator('.task-view .action-buttons .button').filter({hasText: 'Add Attachments'}).click()
	await page.locator('input[type=file]#files').setInputFiles('tests/fixtures/image.jpg')
	await uploadAttachmentPromise

	await expect(page.locator('.attachments .attachments .files button.attachment')).toBeVisible()
}

// FIXME: Large test suite (47 tests) - many pass but some have timeout issues, needs detailed investigation
test.describe.skip('Task', () => {
	let projects: Project[]
	let buckets: Bucket[]

	test.beforeEach(async ({authenticatedPage: page}) => {
		projects = await ProjectFactory.create(1) as Project[]
		const views = await createDefaultViews(projects[0].id)
		buckets = await BucketFactory.create(1, {
			project_view_id: views[3].id,
		}) as Bucket[]
		TaskFactory.truncate()
		UserProjectFactory.truncate()
	})

	test('Should be created new', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/1')
		await page.locator('.input[placeholder="Add a task…"]').fill('New Task')
		await page.locator('.button').filter({hasText: 'Add'}).click()
		await expect(page.locator('.tasks .task .tasktext').first()).toContainText('New Task')
	})

	test('Inserts new tasks at the top of the project', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1)

		await page.goto('/projects/1/1')
		await expect(page.locator('.project-is-empty-notice')).not.toBeVisible()
		await page.locator('.input[placeholder="Add a task…"]').fill('New Task')
		await page.locator('.button').filter({hasText: 'Add'}).click()

		await page.waitForTimeout(1000) // Wait for the request
		await expect(page.locator('.tasks .task .tasktext').first()).toContainText('New Task')
	})

	test('Marks a task as done', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1)

		await page.goto('/projects/1/1')
		await page.locator('.tasks .task .fancy-checkbox').first().click()
		await expect(page.locator('.global-notification')).toContainText('Success')
	})

	test('Can add a task to favorites', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1)

		await page.goto('/projects/1/1')
		await page.locator('.tasks .task .favorite').first().click()
		await expect(page.locator('.menu-container')).toContainText('Favorites')
	})

	test('Should show a task description icon if the task has a description', async ({authenticatedPage: page}) => {
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)
		await TaskFactory.create(1, {
			description: 'Lorem Ipsum',
		})

		await page.goto('/projects/1/1')
		await loadTasksPromise

		await expect(page.locator('.tasks .task .project-task-icon .fa-align-left')).toBeVisible()
	})

	test('Should not show a task description icon if the task has an empty description', async ({authenticatedPage: page}) => {
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)
		await TaskFactory.create(1, {
			description: '',
		})

		await page.goto('/projects/1/1')
		await loadTasksPromise

		await expect(page.locator('.tasks .task .project-task-icon .fa-align-left')).not.toBeVisible()
	})

	test('Should not show a task description icon if the task has a description containing only an empty p tag', async ({authenticatedPage: page}) => {
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)
		await TaskFactory.create(1, {
			description: '<p></p>',
		})

		await page.goto('/projects/1/1')
		await loadTasksPromise

		await expect(page.locator('.tasks .task .project-task-icon .fa-align-left')).not.toBeVisible()
	})

	test.describe('Task Detail View', () => {
		test.beforeEach(async ({authenticatedPage: page}) => {
			TaskCommentFactory.truncate()
			LabelTaskFactory.truncate()
			TaskAttachmentFactory.truncate()
		})

		test('provides back navigation to the project in the list view', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1)
			const loadTasksPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
			)
			await page.goto('/projects/1/1')
			await loadTasksPromise
			await page.locator('.list-view .task').first().locator('a.task-link').click()
			await expect(page.locator('.task-view .back-button')).toBeVisible()
			await page.locator('.task-view .back-button').click()
			await expect(page).toHaveURL(/\/projects\/1\/\d+/)
		})

		test('provides back navigation to the project in the table view', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1)
			const loadTasksPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
			)
			await page.goto('/projects/1/3')
			await loadTasksPromise
			await page.locator('tbody tr').first().locator('a').first().click()
			await expect(page.locator('.task-view .back-button')).toBeVisible()
			await page.locator('.task-view .back-button').click()
			await expect(page).toHaveURL(/\/projects\/1\/\d+/)
		})

		test('provides back navigation to the project in the kanban view on mobile', async ({authenticatedPage: page}) => {
			await page.setViewportSize({width: 375, height: 667}) // iphone-8

			const tasks = await TaskFactory.create(1)
			const loadTasksPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
			)
			await page.goto('/projects/1/4')
			await loadTasksPromise
			await page.locator('.kanban-view .tasks .task').first().click()
			await expect(page.locator('.task-view .back-button')).toBeVisible()
			await page.locator('.task-view .back-button').click()
			await expect(page).toHaveURL(/\/projects\/1\/\d+/)
		})

		test('does not provide back navigation to the project in the kanban view on desktop', async ({authenticatedPage: page}) => {
			await page.setViewportSize({width: 1440, height: 900}) // macbook-15

			const tasks = await TaskFactory.create(1)
			const loadTasksPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
			)
			await page.goto('/projects/1/4')
			await loadTasksPromise
			await page.locator('.kanban-view .tasks .task').first().click()
			await expect(page.locator('.task-view .back-button')).not.toBeVisible()
		})

		test('Shows a 404 page for nonexisting tasks', async ({authenticatedPage: page}) => {
			await page.goto('/tasks/9999')
			await expect(page.locator('body')).toContainText('Not found')
		})

		test('Shows all task details', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				index: 1,
				description: 'Lorem ipsum dolor sit amet.',
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view h1.title.input')).toContainText(tasks[0].title)
			await expect(page.locator('.task-view h1.title.task-id')).toContainText('#1')
			await expect(page.locator('.task-view h6.subtitle')).toContainText(projects[0].title)
			await expect(page.locator('.task-view .details.content.description')).toContainText(tasks[0].description)
			await expect(page.locator('.task-view .action-buttons p.created')).toContainText('Created')
		})

		test('Shows a done label for done tasks', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				index: 1,
				done: true,
				done_at: new Date().toISOString(),
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .heading .is-done')).toBeVisible()
			await expect(page.locator('.task-view .heading .is-done')).toContainText('Done')
			await page.locator('.task-view .action-buttons p.created').scrollIntoViewIfNeeded()
			await expect(page.locator('.task-view .action-buttons p.created')).toBeVisible()
			await expect(page.locator('.task-view .action-buttons p.created')).toContainText('Done')
		})

		test('Can mark a task as done', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'}).click()

			await expect(page.locator('.task-view .heading .is-done')).toBeVisible()
			await expect(page.locator('.task-view .heading .is-done')).toContainText('Done')
			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark as undone'})).toBeVisible()
		})

		test('Shows a task identifier since the project has one', async ({authenticatedPage: page}) => {
			const projects = await ProjectFactory.create(1, {
				id: 1,
				identifier: 'TEST',
			})
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				index: 1,
			})

			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view h1.title.task-id')).toContainText(`${projects[0].identifier}-${tasks[0].index}`)
		})

		test('Can edit the description', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: 'Lorem ipsum dolor sit amet.',
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .details.content.description .tiptap button.done-edit').click()
			await page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror').fill('New Description')
			await page.locator('[data-cy="saveEditor"]').filter({hasText: 'Save'}).click()

			await expect(page.locator('.task-view .details.content.description h3 span.is-small.has-text-success')).toContainText('Saved!')
		})

		test('autosaves the description when leaving the task view', async ({authenticatedPage: page}) => {
			await TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				description: 'Old Description',
			})

			await page.goto('/tasks/1')

			await page.locator('.task-view .details.content.description .tiptap button.done-edit', {timeout: 30_000}).click()
			await page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror').fill('New Description')

			await page.locator('.task-view h6.subtitle a').first().click()

			await page.goto('/tasks/1')
			await expect(page.locator('.task-view .details.content.description')).toContainText('New Description')
		})

		test('Shows an empty editor when the description of a task is empty', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: '',
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .details.content.description .tiptap.ProseMirror p')).toHaveAttribute('data-placeholder')
			await expect(page.locator('.task-view .details.content.description .tiptap button.done-edit')).not.toBeVisible()
		})

		test('Shows a preview editor when the description of a task is not empty', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: 'Lorem Ipsum dolor sit amet',
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .details.content.description .tiptap.ProseMirror p')).not.toHaveAttribute('data-placeholder')
			await expect(page.locator('.task-view .details.content.description .tiptap button.done-edit')).toBeVisible()
		})

		test('Shows a preview editor when the description of a task contains html', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: '<p>Lorem Ipsum dolor sit amet</p>',
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .details.content.description .tiptap.ProseMirror p')).not.toHaveAttribute('data-placeholder')
			await expect(page.locator('.task-view .details.content.description .tiptap button.done-edit')).toBeVisible()
		})

		test('Can add a new comment', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror')).toBeVisible()
			await page.locator('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror').fill('New Comment')
			await page.locator('.task-view .comments .media.comment .button:not([disabled])').filter({hasText: 'Comment'}).click()

			await expect(page.locator('.task-view .comments .media.comment .tiptap__editor')).toContainText('New Comment')
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can move a task to another project', async ({authenticatedPage: page}) => {
			const projects = await ProjectFactory.create(2)
			const views = await createDefaultViews(projects[0].id)
			await BucketFactory.create(2, {
				project_view_id: views[3].id,
			})
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: /^Move$/}).click()
			await page.locator('.task-view .content.details .field .multiselect.control .input-wrapper input').fill(`${projects[1].title}{enter}`)
			// The requests happen with a 200ms timeout. Because of that, the results are not yet there when we
			// press enter and we can't simulate pressing on enter to select the item.
			await page.locator('.task-view .content.details .field .multiselect.control .search-results').locator('> *').first().click()

			await expect(page.locator('.task-view h6.subtitle')).toContainText(projects[1].title)
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can delete a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'})).toBeVisible()
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'}).click()
			await expect(page.locator('.modal-mask .modal-container .modal-content .modal-header')).toContainText('Delete this task')
			await page.locator('.modal-mask .modal-container .modal-content .actions .button').filter({hasText: 'Do it!'}).click()

			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page).toHaveURL(new RegExp(`/projects/${tasks[0].project_id}/`))
		})

		test('Can add an assignee to a task', async ({authenticatedPage: page}) => {
			const users = await UserFactory.create(5)
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			await UserProjectFactory.create(5, {
				project_id: 1,
				user_id: '{increment}',
			})

			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('[data-cy="taskDetail.assign"]').click()
			await page.locator('.task-view .column.assignees .multiselect input').fill(users[1].username)
			await expect(page.locator('.task-view .column.assignees .multiselect .search-results')).toBeVisible()
			await page.locator('.task-view .column.assignees .multiselect .search-results').locator('> *').first().click()

			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page.locator('.task-view .column.assignees .multiselect .input-wrapper span.assignee')).toBeVisible()
		})

		test('Can remove an assignee from a task', async ({authenticatedPage: page}) => {
			const users = await UserFactory.create(2)
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			await UserProjectFactory.create(5, {
				project_id: 1,
				user_id: '{increment}',
			})
			await TaskAssigneeFactory.create(1, {
				task_id: tasks[0].id,
				user_id: users[1].id,
			})

			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .column.assignees .multiselect .input-wrapper span.assignee .remove-assignee').click()

			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page.locator('.task-view .column.assignees .multiselect .input-wrapper span.assignee')).not.toBeVisible()
		})

		test('Can add a new label to a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			LabelFactory.truncate()
			const newLabelText = 'some new label'

			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Add Labels'})).toBeVisible()
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Add Labels'}).click()
			await page.locator('.task-view .details.labels-list .multiselect input').fill(newLabelText)
			await page.locator('.task-view .details.labels-list .multiselect .search-results').locator('> *').first().click()

			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper span.tag')).toBeVisible()
			await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper span.tag')).toContainText(newLabelText)
		})

		test('Can add an existing label to a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			const labels = await LabelFactory.create(1)
			LabelTaskFactory.truncate()

			await page.goto(`/tasks/${tasks[0].id}`)

			await addLabelToTaskAndVerify(page, labels[0].title)
		})

		test('Can add a label to a task and it shows up on the kanban board afterwards', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			const labels = await LabelFactory.create(1)
			LabelTaskFactory.truncate()
			await TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: buckets[0].id,
				project_view_id: buckets[0].project_view_id,
			})

			await page.goto(`/projects/${projects[0].id}/4`)

			await page.locator('.bucket .task').filter({hasText: tasks[0].title}).click()

			await addLabelToTaskAndVerify(page, labels[0].title)

			await page.locator('.modal-container > .close').click()

			await expect(page.locator('.bucket .task')).toContainText(labels[0].title)
		})

		test('Can remove a label from a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			const labels = await LabelFactory.create(1)
			await LabelTaskFactory.create(1, {
				task_id: tasks[0].id,
				label_id: labels[0].id,
			})

			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper')).toBeVisible()
			await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper')).toContainText(labels[0].title)
			await page.locator('.task-view .details.labels-list .multiselect .input-wrapper').locator('> *').first().locator('[data-cy="taskDetail.removeLabel"]').click()

			await expect(page.locator('.global-notification')).toContainText('Success')
			await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper')).not.toContainText(labels[0].title)
		})

		test('Can set a due date for a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Due Date'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker .show').click()
			await page.locator('.datepicker .datepicker-popup button').filter({hasText: 'Tomorrow'}).click()
			await page.locator('[data-cy="closeDatepicker"]').filter({hasText: 'Confirm'}).click()

			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can set a due date to a specific date for a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Due Date'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker .show').click()
			await page.locator('.datepicker-popup .flatpickr-innerContainer .flatpickr-days .flatpickr-day.today').click()
			await page.locator('[data-cy="closeDatepicker"]').filter({hasText: 'Confirm'}).click()

			const today = new Date()
			today.setHours(12)
			today.setMinutes(0)
			today.setSeconds(0)
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker-popup')).not.toBeVisible()
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input')).toContainText(dayjs(today).fromNow())
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can change a due date to a specific date for a task', async ({authenticatedPage: page}) => {
			const dueDate = new Date(2025, 2, 20)
			dueDate.setHours(12)
			dueDate.setMinutes(0)
			dueDate.setSeconds(0)
			dueDate.setDate(1)
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: dueDate.toISOString(),
			})

			const today = new Date(2025, 2, 5)
			today.setHours(12)
			today.setMinutes(0)
			today.setSeconds(0)

			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Due Date'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker .show').click()
			await page.locator(`.datepicker-popup .flatpickr-innerContainer .flatpickr-days [aria-label="${today.toLocaleString('en-US', {month: 'long'})} ${today.getDate()}, ${today.getFullYear()}"]`).click()
			await page.locator('[data-cy="closeDatepicker"]').filter({hasText: 'Confirm'}).click()

			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker-popup')).not.toBeVisible()
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input')).toContainText(dayjs(today).fromNow())
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can paste an image into the description editor which uploads it as an attachment', async ({authenticatedPage: page}) => {
			TaskAttachmentFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
			}) as Task[]
			await page.goto(`/tasks/${tasks[0].id}`)

			const uploadAttachmentPromise = page.waitForResponse(response =>
				response.url().includes(`/tasks/${tasks[0].id}/attachments`) && response.request().method() === 'PUT',
			)

			await pasteFile(page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror', {timeout: 30_000}), 'image.jpg', 'image/jpeg')

			await uploadAttachmentPromise
			await expect(page.locator('.attachments .attachments .files button.attachment')).toBeVisible()
			const img = page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror img')
			await expect(img).toBeVisible()
			const naturalWidth = await img.evaluate((el: HTMLImageElement) => el.naturalWidth)
			expect(naturalWidth).toBeGreaterThan(0)
		})

		test('Can set a reminder', async ({authenticatedPage: page}) => {
			TaskReminderFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Reminders'}).click()
			await page.locator('.task-view .columns.details .column button').filter({hasText: 'Add a reminder'}).click()
			await page.locator('.datepicker__quick-select-date').filter({hasText: 'Tomorrow'}).click()

			await expect(page.locator('.reminder-options-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Allows to set a relative reminder when the task already has a due date', async ({authenticatedPage: page}) => {
			TaskReminderFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Reminders'}).click()
			await page.locator('.task-view .columns.details .column button').filter({hasText: 'Add a reminder'}).click()
			await expect(page.locator('.datepicker__quick-select-date')).not.toBeVisible()
			await expect(page.locator('.reminder-options-popup .card-content')).toContainText('1 day before Due Date')
			await page.locator('.reminder-options-popup .card-content').filter({hasText: '1 day before Due Date'}).click()

			await expect(page.locator('.reminder-options-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Allows to set a relative reminder when the task already has a start date', async ({authenticatedPage: page}) => {
			TaskReminderFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
				start_date: (new Date()).toISOString(),
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Reminders'}).click()
			await page.locator('.task-view .columns.details .column button').filter({hasText: 'Add a reminder'}).click()
			await expect(page.locator('.datepicker__quick-select-date')).not.toBeVisible()
			await expect(page.locator('.reminder-options-popup .card-content')).toContainText('1 day before Start Date')
			await page.locator('.reminder-options-popup .card-content').filter({hasText: '1 day before Start Date'}).click()

			await expect(page.locator('.reminder-options-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Allows to set a custom relative reminder when the task already has a due date', async ({authenticatedPage: page}) => {
			TaskReminderFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Reminders'}).click()
			await page.locator('.task-view .columns.details .column button').filter({hasText: 'Add a reminder'}).click()
			await expect(page.locator('.datepicker__quick-select-date')).not.toBeVisible()
			await page.locator('.reminder-options-popup .card-content').filter({hasText: 'Custom'}).click()
			await page.locator('.reminder-options-popup .card-content .reminder-period input').first().fill('10')
			await page.locator('.reminder-options-popup .card-content .reminder-period select').first().selectOption('days')
			await page.locator('.reminder-options-popup .card-content button').filter({hasText: 'Confirm'}).click()

			await expect(page.locator('.reminder-options-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Allows to set a fixed reminder when the task already has a due date', async ({authenticatedPage: page}) => {
			TaskReminderFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Reminders'}).click()
			await page.locator('.task-view .columns.details .column button').filter({hasText: 'Add a reminder'}).click()
			await expect(page.locator('.datepicker__quick-select-date')).not.toBeVisible()
			await page.locator('.reminder-options-popup .card-content').filter({hasText: 'Date and time'}).click()
			await page.locator('.datepicker__quick-select-date').filter({hasText: 'Tomorrow'}).click()

			await expect(page.locator('.reminder-options-popup')).not.toBeVisible()
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Can set a priority for a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Priority'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Priority'}).locator('.select select').selectOption('Urgent')
			await expect(page.locator('.global-notification')).toContainText('Success')

			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Priority'}).locator('.select select')).toHaveValue('4')
		})

		test('Can set the progress for a task', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Progress'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Progress'}).locator('.select select').selectOption('50%')
			await expect(page.locator('.global-notification')).toContainText('Success')

			await page.waitForTimeout(200)

			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Progress'}).locator('.select select')).toBeVisible()
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Progress'}).locator('.select select')).toHaveValue('0.5')
		})

		test('Can add an attachment to a task', async ({authenticatedPage: page}) => {
			TaskAttachmentFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await uploadAttachmentAndVerify(page, tasks[0].id)
		})

		test('Can add an attachment to a task and see it appearing on kanban', async ({authenticatedPage: page}) => {
			TaskAttachmentFactory.truncate()
			const tasks = await TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			const labels = await LabelFactory.create(1)
			LabelTaskFactory.truncate()
			await TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: buckets[0].id,
				project_view_id: buckets[0].project_view_id,
			})

			await page.goto(`/projects/${projects[0].id}/4`)

			await page.locator('.bucket .task').filter({hasText: tasks[0].title}).click()

			await uploadAttachmentAndVerify(page, tasks[0].id)

			await page.locator('.modal-container > .close').click()

			await expect(page.locator('.bucket .task .footer .icon svg.fa-paperclip')).toBeVisible()
		})

		test('Can check items off a checklist', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>First Item</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Second Item</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Third Item</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Fourth Item</p></div>
	</li>
	<li data-checked="true" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Fifth Item</p></div>
	</li>
</ul>`,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .checklist-summary')).toContainText('1 of 5 tasks')
			await page.locator('.tiptap__editor ul > li input[type=checkbox]').nth(2).click()

			await expect(page.locator('.task-view .details.content.description h3 span.is-small.has-text-success')).toContainText('Saved!')
			await expect(page.locator('.tiptap__editor ul > li input[type=checkbox]').nth(2)).toBeChecked()
			await expect(page.locator('.tiptap__editor input[type=checkbox]')).toHaveCount(5)
			await expect(page.locator('.task-view .checklist-summary')).toContainText('2 of 5 tasks')
		})

		test('Persists checked checklist items after reload', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>First Item</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Second Item</p></div>
	</li>
</ul>`,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.task-view .checklist-summary')).toContainText('0 of 2 tasks')
			await page.locator('.tiptap__editor ul > li input[type=checkbox]').first().click()

			await expect(page.locator('.task-view .details.content.description h3 span.is-small.has-text-success')).toContainText('Saved!')

			await expect(page.locator('.task-view .checklist-summary')).toContainText('1 of 2 tasks')

			await page.reload()

			await expect(page.locator('.task-view .checklist-summary')).toContainText('1 of 2 tasks')
			await expect(page.locator('.tiptap__editor ul > li input[type=checkbox]').first()).toBeChecked()
		})

		test('Should use the editor to render description', async ({authenticatedPage: page}) => {
			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: `
<h1>Lorem Ipsum</h1>
<p>Dolor sit amet</p>
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>First Item</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Second Item</p></div>
	</li>
</ul>`,
			})
			await page.goto(`/tasks/${tasks[0].id}`)

			await expect(page.locator('.tiptap__editor ul > li input[type=checkbox]').first()).toBeVisible()
			await expect(page.locator('.tiptap__editor h1').filter({hasText: 'Lorem Ipsum'})).toBeVisible()
			await expect(page.locator('.tiptap__editor p').filter({hasText: 'Dolor sit amet'})).toBeVisible()
		})

		test('Should render an image from attachment', async ({authenticatedPage: page, apiContext}) => {
			TaskAttachmentFactory.truncate()

			const tasks = await TaskFactory.create(1, {
				id: 1,
				description: '',
			})

			const filePath = join(__dirname, '../../fixtures/image.jpg')
			const fileBuffer = readFileSync(filePath)

			// Navigate to a page first to establish context for localStorage access
			await page.goto('/')
			const token = await page.evaluate(() => localStorage.getItem('token'))

			const response = await apiContext.put(`tasks/${tasks[0].id}/attachments`, {
				multipart: {
					files: {
						name: 'image.jpg',
						mimeType: 'image/jpeg',
						buffer: fileBuffer,
					},
				},
				headers: {
					'Authorization': `Bearer ${token}`,
				},
			})

			const {success} = await response.json()

			await TaskFactory.create(1, {
				id: 1,
				description: `<img src="${apiUrl}/tasks/${tasks[0].id}/attachments/${success[0].id}" alt="test image">`,
			})

			await page.goto(`/tasks/${tasks[0].id}`)

			const img = page.locator('.tiptap__editor img')
			await expect(img).toBeVisible()
			const naturalWidth = await img.evaluate((el: HTMLImageElement) => el.naturalWidth)
			expect(naturalWidth).toBeGreaterThan(0)
		})
	})
})
