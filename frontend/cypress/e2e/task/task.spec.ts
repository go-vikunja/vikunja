import {createFakeUserAndLogin} from '../../support/authenticateUser'

import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

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

function addLabelToTaskAndVerify(labelTitle: string) {
	cy.get('.task-view .action-buttons .button')
		.contains('Add Labels')
		.click()
	cy.get('.task-view .details.labels-list .multiselect input')
		.type(labelTitle)
	cy.get('.task-view .details.labels-list .multiselect .search-results')
		.children()
		.first()
		.click()

	cy.get('.global-notification', {timeout: 4000})
		.should('contain', 'Success')
	cy.get('.task-view .details.labels-list .multiselect .input-wrapper span.tag')
		.should('exist')
		.should('contain', labelTitle)
}

function uploadAttachmentAndVerify(taskId: number) {
	cy.intercept(`**/tasks/${taskId}/attachments`).as('uploadAttachment')
	cy.get('.task-view .action-buttons .button')
		.contains('Add Attachments')
		.click()
	cy.get('input[type=file]#files', {timeout: 1000})
		.selectFile('cypress/fixtures/image.jpg', {force: true}) // The input is not visible, but on purpose
	cy.wait('@uploadAttachment')

	cy.get('.attachments .attachments .files button.attachment')
		.should('exist')
}

describe('Task', () => {
	createFakeUserAndLogin()

	let projects: Project[]
	let buckets: Bucket[]

	beforeEach(() => {
		// UserFactory.create(1)
		projects = ProjectFactory.create(1) as Project[]
		const views = createDefaultViews(projects[0].id)
		buckets = BucketFactory.create(1, {
			project_view_id: views[3].id,
		}) as Bucket[]
		TaskFactory.truncate()
		UserProjectFactory.truncate()
	})

	it('Should be created new', () => {
		cy.visit('/projects/1/1')
		cy.get('.input[placeholder="Add a task…"]')
			.type('New Task')
		cy.get('.button')
			.contains('Add')
			.click()
		cy.get('.tasks .task .tasktext')
			.first()
			.should('contain', 'New Task')
	})

	it('Inserts new tasks at the top of the project', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/1')
		cy.get('.project-is-empty-notice')
			.should('not.exist')
		cy.get('.input[placeholder="Add a task…"]')
			.type('New Task')
		cy.get('.button')
			.contains('Add')
			.click()

		cy.wait(1000) // Wait for the request
		cy.get('.tasks .task .tasktext')
			.first()
			.should('contain', 'New Task')
	})

	it('Marks a task as done', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/1')
		cy.get('.tasks .task .fancy-checkbox')
			.first()
			.click()
		cy.get('.global-notification')
			.should('contain', 'Success')
	})

	it('Can add a task to favorites', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/1')
		cy.get('.tasks .task .favorite')
			.first()
			.click()
		cy.get('.menu-container')
			.should('contain', 'Favorites')
	})

	it('Should show a task description icon if the task has a description', () => {
		cy.intercept('**/projects/1/views/*/tasks**').as('loadTasks')
		TaskFactory.create(1, {
			description: 'Lorem Ipsum',
		})

		cy.visit('/projects/1/1')
		cy.wait('@loadTasks')

		cy.get('.tasks .task .project-task-icon .fa-align-left')
			.should('exist')
	})

	it('Should not show a task description icon if the task has an empty description', () => {
		cy.intercept('**/projects/1/views/*/tasks**').as('loadTasks')
		TaskFactory.create(1, {
			description: '',
		})

		cy.visit('/projects/1/1')
		cy.wait('@loadTasks')

		cy.get('.tasks .task .project-task-icon .fa-align-left')
			.should('not.exist')
	})

	it('Should not show a task description icon if the task has a description containing only an empty p tag', () => {
		cy.intercept('**/projects/1/views/*/tasks**').as('loadTasks')
		TaskFactory.create(1, {
			description: '<p></p>',
		})

		cy.visit('/projects/1/1')
		cy.wait('@loadTasks')

		cy.get('.tasks .task .project-task-icon .fa-align-left')
			.should('not.exist')
	})

	describe('Task Detail View', () => {
		beforeEach(() => {
			TaskCommentFactory.truncate()
			LabelTaskFactory.truncate()
			TaskAttachmentFactory.truncate()
		})
		it('Shows a 404 page for nonexisting tasks', () => {

			cy.visit('/tasks/9999')

			cy.contains('Not found')
				.should('be.visible')
		})
		it('Shows all task details', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				index: 1,
				description: 'Lorem ipsum dolor sit amet.',
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view h1.title.input')
				.should('contain', tasks[0].title)
			cy.get('.task-view h1.title.task-id')
				.should('contain', '#1')
			cy.get('.task-view h6.subtitle')
				.should('contain', projects[0].title)
			cy.get('.task-view .details.content.description')
				.should('contain', tasks[0].description)
			cy.get('.task-view .action-buttons p.created')
				.should('contain', 'Created')
		})

		it('Shows a done label for done tasks', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				index: 1,
				done: true,
				done_at: new Date().toISOString(),
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .heading .is-done')
				.should('be.visible')
				.should('contain', 'Done')
			cy.get('.task-view .action-buttons p.created')
				.scrollIntoView()
				.should('be.visible')
				.should('contain', 'Done')
		})

		it('Can mark a task as done', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Mark task done!')
				.click()

			cy.get('.task-view .heading .is-done')
				.should('exist')
				.should('contain', 'Done')
			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.get('.task-view .action-buttons .button')
				.should('contain', 'Mark as undone')
		})

		it('Shows a task identifier since the project has one', () => {
			const projects = ProjectFactory.create(1, {
				id: 1,
				identifier: 'TEST',
			})
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				index: 1,
			})

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view h1.title.task-id')
				.should('contain', `${projects[0].identifier}-${tasks[0].index}`)
		})

		it('Can edit the description', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				description: 'Lorem ipsum dolor sit amet.',
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.content.description .tiptap button.done-edit')
				.click()
			cy.get('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror')
				.type('{selectall}New Description')
			cy.get('[data-cy="saveEditor"]')
				.contains('Save')
				.click()

			cy.get('.task-view .details.content.description h3 span.is-small.has-text-success')
				.contains('Saved!')
				.should('exist')
		})

		it('autosaves the description when leaving the task view', () => {
			TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
				description: 'Old Description',
			})

			cy.visit('/tasks/1')
			
			cy.get('.task-view .details.content.description .tiptap button.done-edit', {timeout: 30_000})
				.click()
			cy.get('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror')
				.type('{selectall}New Description')
			
			cy.get('.task-view h6.subtitle a')
				.first()
				.click()
			
			cy.visit('/tasks/1')
			cy.get('.task-view .details.content.description')
				.should('contain.text', 'New Description')
		})

		it('Shows an empty editor when the description of a task is empty', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				description: '',
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.content.description .tiptap.ProseMirror p')
				.should('have.attr', 'data-placeholder')
			cy.get('.task-view .details.content.description .tiptap button.done-edit')
				.should('not.exist')
		})

		it('Shows a preview editor when the description of a task is not empty', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				description: 'Lorem Ipsum dolor sit amet',
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.content.description .tiptap.ProseMirror p')
				.should('not.have.attr', 'data-placeholder')
			cy.get('.task-view .details.content.description .tiptap button.done-edit')
				.should('exist')
		})

		it('Shows a preview editor when the description of a task contains html', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				description: '<p>Lorem Ipsum dolor sit amet</p>',
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.content.description .tiptap.ProseMirror p')
				.should('not.have.attr', 'data-placeholder')
			cy.get('.task-view .details.content.description .tiptap button.done-edit')
				.should('exist')
		})

		it('Can add a new comment', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .comments .media.comment .tiptap__editor .tiptap.ProseMirror')
				.should('be.visible')
				.type('{selectall}New Comment')
			cy.get('.task-view .comments .media.comment .button:not([disabled])')
				.contains('Comment')
				.should('be.visible')
				.click()

			cy.get('.task-view .comments .media.comment .tiptap__editor')
				.should('contain', 'New Comment')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can move a task to another project', () => {
			const projects = ProjectFactory.create(2)
			const views = createDefaultViews(projects[0].id)
			BucketFactory.create(2, {
				project_view_id: views[3].id,
			})
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Move')
				.click()
			cy.get('.task-view .content.details .field .multiselect.control .input-wrapper input')
				.type(`${projects[1].title}{enter}`)
			// The requests happen with a 200ms timeout. Because of that, the results are not yet there when cypress 
			// presses enter and we can't simulate pressing on enter to select the item.
			cy.get('.task-view .content.details .field .multiselect.control .search-results')
				.children()
				.first()
				.click()

			cy.get('.task-view h6.subtitle')
				.should('contain', projects[1].title)
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can delete a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.should('be.visible')
				.contains('Delete')
				.click()
			cy.get('.modal-mask .modal-container .modal-content .modal-header')
				.should('contain', 'Delete this task')
			cy.get('.modal-mask .modal-container .modal-content .actions .button')
				.contains('Do it!')
				.click()

			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.url()
				.should('contain', `/projects/${tasks[0].project_id}/`)
		})

		it('Can add an assignee to a task', () => {
			const users = UserFactory.create(5)
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			UserProjectFactory.create(5, {
				project_id: 1,
				user_id: '{increment}',
			})

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('[data-cy="taskDetail.assign"]')
				.click()
			cy.get('.task-view .column.assignees .multiselect input')
				.type(users[1].username)
			cy.get('.task-view .column.assignees .multiselect .search-results')
				.should('be.visible')
				.children()
				.first()
				.click()

			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.get('.task-view .column.assignees .multiselect .input-wrapper span.assignee')
				.should('exist')
		})

		it('Can remove an assignee from a task', () => {
			const users = UserFactory.create(2)
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			UserProjectFactory.create(5, {
				project_id: 1,
				user_id: '{increment}',
			})
			TaskAssigneeFactory.create(1, {
				task_id: tasks[0].id,
				user_id: users[1].id,
			})

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .column.assignees .multiselect .input-wrapper span.assignee')
				.get('.remove-assignee')
				.click()

			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.get('.task-view .column.assignees .multiselect .input-wrapper span.assignee')
				.should('not.exist')
		})

		it('Can add a new label to a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			LabelFactory.truncate()
			const newLabelText = 'some new label'

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Add Labels')
				.should('be.visible')
				.click()
			cy.get('.task-view .details.labels-list .multiselect input')
				.type(newLabelText)
			cy.get('.task-view .details.labels-list .multiselect .search-results')
				.children()
				.first()
				.click()

			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.get('.task-view .details.labels-list .multiselect .input-wrapper span.tag')
				.should('exist')
				.should('contain', newLabelText)
		})

		it('Can add an existing label to a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.truncate()

			cy.visit(`/tasks/${tasks[0].id}`)

			addLabelToTaskAndVerify(labels[0].title)
		})

		it('Can add a label to a task and it shows up on the kanban board afterwards', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.truncate()
			TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: buckets[0].id,
				project_view_id: buckets[0].project_view_id,
			})

			cy.visit(`/projects/${projects[0].id}/4`)

			cy.get('.bucket .task')
				.contains(tasks[0].title)
				.click()

			addLabelToTaskAndVerify(labels[0].title)

			cy.get('.modal-container > .close')
				.click()

			cy.get('.bucket .task')
				.should('contain.text', labels[0].title)
		})

		it('Can remove a label from a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: 1,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.create(1, {
				task_id: tasks[0].id,
				label_id: labels[0].id,
			})

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.labels-list .multiselect .input-wrapper')
				.should('be.visible')
				.should('contain', labels[0].title)
			cy.get('.task-view .details.labels-list .multiselect .input-wrapper')
				.children()
				.first()
				.get('[data-cy="taskDetail.removeLabel"]')
				.click()

			cy.get('.global-notification')
				.should('contain', 'Success')
			cy.get('.task-view .details.labels-list .multiselect .input-wrapper')
				.should('not.contain', labels[0].title)
		})

		it('Can set a due date for a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Due Date')
				.click()
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker .show')
				.click()
			cy.get('.datepicker .datepicker-popup button')
				.contains('Tomorrow')
				.click()
			cy.get('[data-cy="closeDatepicker"]')
				.contains('Confirm')
				.click()

			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker-popup')
				.should('not.exist')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can set a due date to a specific date for a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Due Date')
				.click()
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker .show')
				.click()
			cy.get('.datepicker-popup .flatpickr-innerContainer .flatpickr-days .flatpickr-day.today')
				.click()
			cy.get('[data-cy="closeDatepicker"]')
				.contains('Confirm')
				.click()

			const today = new Date()
			today.setHours(12)
			today.setMinutes(0)
			today.setSeconds(0)
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker-popup')
				.should('not.exist')
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input')
				.should('contain.text', dayjs(today).fromNow())
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can change a due date to a specific date for a task', () => {
			const dueDate = new Date(2025, 2, 20)
			dueDate.setHours(12)
			dueDate.setMinutes(0)
			dueDate.setSeconds(0)
			dueDate.setDate(1)
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: dueDate.toISOString(),
			})

			const today = new Date(2025, 2, 5)
			today.setHours(12)
			today.setMinutes(0)
			today.setSeconds(0)

			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Due Date')
				.click()
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker .show')
				.click()
			cy.get(`.datepicker-popup .flatpickr-innerContainer .flatpickr-days [aria-label="${today.toLocaleString('en-US', {month: 'long'})} ${today.getDate()}, ${today.getFullYear()}"]`)
				.click()
			cy.get('[data-cy="closeDatepicker"]')
				.contains('Confirm')
				.click()

			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input .datepicker-popup')
				.should('not.exist')
			cy.get('.task-view .columns.details .column')
				.contains('Due Date')
				.get('.date-input')
				.should('contain.text', dayjs(today).fromNow())
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can paste an image into the description editor which uploads it as an attachment', () => {
			TaskAttachmentFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
			}) as Task[]
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.intercept('**/tasks/*/attachments').as('uploadAttachment')
			
			cy.get('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror', {timeout: 30_000})
				.pasteFile('image.jpg', 'image/jpeg')

			cy.wait('@uploadAttachment')
			cy.get('.attachments .attachments .files button.attachment')
				.should('exist')
			cy.get('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror img')
				.should('be.visible')
				.and(($img) => {
					// "naturalWidth" and "naturalHeight" are set when the image loads
					expect($img[0].naturalWidth).to.be.greaterThan(0)
				})
		})

		it('Can set a reminder', () => {
			TaskReminderFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Reminders')
				.click()
			cy.get('.task-view .columns.details .column button')
				.contains('Add a reminder')
				.click()
			cy.get('.datepicker__quick-select-date')
				.contains('Tomorrow')
				.click()

			cy.get('.reminder-options-popup')
				.should('not.be.visible')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Allows to set a relative reminder when the task already has a due date', () => {
			TaskReminderFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Reminders')
				.click()
			cy.get('.task-view .columns.details .column button')
				.contains('Add a reminder')
				.click()
			cy.get('.datepicker__quick-select-date')
				.should('not.exist')
			cy.get('.reminder-options-popup .card-content')
				.should('contain', '1 day before Due Date')
			cy.get('.reminder-options-popup .card-content')
				.contains('1 day before Due Date')
				.click()

			cy.get('.reminder-options-popup')
				.should('not.be.visible')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Allows to set a relative reminder when the task already has a start date', () => {
			TaskReminderFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
				start_date: (new Date()).toISOString(),
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Reminders')
				.click()
			cy.get('.task-view .columns.details .column button')
				.contains('Add a reminder')
				.click()
			cy.get('.datepicker__quick-select-date')
				.should('not.exist')
			cy.get('.reminder-options-popup .card-content')
				.should('contain', '1 day before Start Date')
			cy.get('.reminder-options-popup .card-content')
				.contains('1 day before Start Date')
				.click()

			cy.get('.reminder-options-popup')
				.should('not.be.visible')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Allows to set a custom relative reminder when the task already has a due date', () => {
			TaskReminderFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Reminders')
				.click()
			cy.get('.task-view .columns.details .column button')
				.contains('Add a reminder')
				.click()
			cy.get('.datepicker__quick-select-date')
				.should('not.exist')
			cy.get('.reminder-options-popup .card-content')
				.contains('Custom')
				.click()
			cy.get('.reminder-options-popup .card-content .reminder-period input')
				.first()
				.type('{selectall}10')
			cy.get('.reminder-options-popup .card-content .reminder-period select')
				.first()
				.select('days')
			cy.get('.reminder-options-popup .card-content button')
				.contains('Confirm')
				.click()

			cy.get('.reminder-options-popup')
				.should('not.be.visible')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Allows to set a fixed reminder when the task already has a due date', () => {
			TaskReminderFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				done: false,
				due_date: (new Date()).toISOString(),
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Reminders')
				.click()
			cy.get('.task-view .columns.details .column button')
				.contains('Add a reminder')
				.click()
			cy.get('.datepicker__quick-select-date')
				.should('not.exist')
			cy.get('.reminder-options-popup .card-content')
				.contains('Date and time')
				.click()
			cy.get('.datepicker__quick-select-date')
				.contains('Tomorrow')
				.click()

			cy.get('.reminder-options-popup')
				.should('not.be.visible')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can set a priority for a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Priority')
				.click()
			cy.get('.task-view .columns.details .column')
				.contains('Priority')
				.get('.select select')
				.select('Urgent')
			cy.get('.global-notification')
				.should('contain', 'Success')

			cy.get('.task-view .columns.details .column')
				.contains('Priority')
				.get('.select select')
				.should('have.value', '4')
		})

		it('Can set the progress for a task', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .action-buttons .button')
				.contains('Set Progress')
				.click()
			cy.get('.task-view .columns.details .column')
				.contains('Progress')
				.get('.select select')
				.select('50%')
			cy.get('.global-notification')
				.should('contain', 'Success')

			cy.wait(200)

			cy.get('.task-view .columns.details .column')
				.contains('Progress')
				.get('.select select')
				.should('be.visible')
				.should('have.value', '0.5')
		})

		it('Can add an attachment to a task', () => {
			TaskAttachmentFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			uploadAttachmentAndVerify(tasks[0].id)
		})

		it('Can add an attachment to a task and see it appearing on kanban', () => {
			TaskAttachmentFactory.truncate()
			const tasks = TaskFactory.create(1, {
				id: 1,
				project_id: projects[0].id,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.truncate()
			TaskBucketFactory.create(1, {
				task_id: tasks[0].id,
				bucket_id: buckets[0].id,
				project_view_id: buckets[0].project_view_id,
			})

			cy.visit(`/projects/${projects[0].id}/4`)

			cy.get('.bucket .task')
				.contains(tasks[0].title)
				.click()

			uploadAttachmentAndVerify(tasks[0].id)

			cy.get('.modal-container > .close')
				.click()

			cy.get('.bucket .task .footer .icon svg.fa-paperclip')
				.should('exist')
		})

		it('Can check items off a checklist', () => {
			const tasks = TaskFactory.create(1, {
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
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .checklist-summary')
				.should('contain.text', '1 of 5 tasks')
			cy.get('.tiptap__editor ul > li input[type=checkbox]')
				.eq(2)
				.click()

			cy.get('.task-view .details.content.description h3 span.is-small.has-text-success')
				.contains('Saved!')
				.should('exist')
			cy.get('.tiptap__editor ul > li input[type=checkbox]')
				.eq(2)
				.should('be.checked')
			cy.get('.tiptap__editor input[type=checkbox]')
				.should('have.length', 5)
			cy.get('.task-view .checklist-summary')
				.should('contain.text', '2 of 5 tasks')
			})

		it('Persists checked checklist items after reload', () => {
			const tasks = TaskFactory.create(1, {
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
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .checklist-summary')
				.should('contain.text', '0 of 2 tasks')
			cy.get('.tiptap__editor ul > li input[type=checkbox]')
				.first()
				.click()

			cy.get('.task-view .details.content.description h3 span.is-small.has-text-success')
				.contains('Saved!')
				.should('exist')

			cy.get('.task-view .checklist-summary')
				.should('contain.text', '1 of 2 tasks')

			cy.reload()

			cy.get('.task-view .checklist-summary')
				.should('contain.text', '1 of 2 tasks')
			cy.get('.tiptap__editor ul > li input[type=checkbox]')
				.first()
				.should('be.checked')
		})

		it('Should use the editor to render description', () => {
			const tasks = TaskFactory.create(1, {
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
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.tiptap__editor ul > li input[type=checkbox]')
				.should('exist')
			cy.get('.tiptap__editor h1')
				.contains('Lorem Ipsum')
				.should('exist')
			cy.get('.tiptap__editor p')
				.contains('Dolor sit amet')
				.should('exist')
		})

		it('Should render an image from attachment', async () => {

			TaskAttachmentFactory.truncate()

			const tasks = TaskFactory.create(1, {
				id: 1,
				description: '',
			})

			cy.readFile('cypress/fixtures/image.jpg', null).then(file => {

				const formData = new FormData()
				formData.append('files', new Blob([file]), 'image.jpg')

				cy.request({
					method: 'PUT',
					url: `${Cypress.env('API_URL')}/tasks/${tasks[0].id}/attachments`,
					headers: {
						'Authorization': `Bearer ${window.localStorage.getItem('token')}`,
						'Content-Type': 'multipart/form-data',
					},
					body: formData,
				})
					.then(({body}) => {
						const dec = new TextDecoder('utf-8')
						const {success} = JSON.parse(dec.decode(body))

						TaskFactory.create(1, {
							id: 1,
							description: `<img src="${Cypress.env('API_URL')}/tasks/${tasks[0].id}/attachments/${success[0].id}" alt="test image">`,
						})

						cy.visit(`/tasks/${tasks[0].id}`)

						cy.get('.tiptap__editor img')
							.should('be.visible')
							.and(($img) => {
								// "naturalWidth" and "naturalHeight" are set when the image loads
								expect($img[0].naturalWidth).to.be.greaterThan(0)
							})

					})
			})
		})
	})
})
