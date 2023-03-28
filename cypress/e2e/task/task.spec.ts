import {createFakeUserAndLogin} from '../../support/authenticateUser'

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

	cy.get('.global-notification', { timeout: 4000 })
		.should('contain', 'Success')
	cy.get('.task-view .details.labels-list .multiselect .input-wrapper span.tag')
		.should('exist')
		.should('contain', labelTitle)
}

function uploadAttachmentAndVerify(taskId: number) {
	cy.intercept(`${Cypress.env('API_URL')}/tasks/${taskId}/attachments`).as('uploadAttachment')
	cy.get('.task-view .action-buttons .button')
		.contains('Add Attachments')
		.click()
	cy.get('input[type=file]', {timeout: 1000})
		.selectFile('cypress/fixtures/image.jpg', {force: true}) // The input is not visible, but on purpose
	cy.wait('@uploadAttachment')

	cy.get('.attachments .attachments .files a.attachment')
		.should('exist')
}

describe('Task', () => {
	createFakeUserAndLogin()

	let projects
	let buckets

	beforeEach(() => {
		// UserFactory.create(1)
		projects = ProjectFactory.create(1)
		buckets = BucketFactory.create(1, {
			project_id: projects[0].id,
		})
		TaskFactory.truncate()
		UserProjectFactory.truncate()
	})

	it('Should be created new', () => {
		cy.visit('/projects/1/list')
		cy.get('.input[placeholder="Add a new task…"')
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

		cy.visit('/projects/1/list')
		cy.get('.project-is-empty-notice')
			.should('not.exist')
		cy.get('.input[placeholder="Add a new task…"')
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

		cy.visit('/projects/1/list')
		cy.get('.tasks .task .fancycheckbox')
			.first()
			.click()
		cy.get('.global-notification')
			.should('contain', 'Success')
	})

	it('Can add a task to favorites', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/list')
		cy.get('.tasks .task .favorite')
			.first()
			.click()
		cy.get('.menu-container')
			.should('contain', 'Favorites')
	})

	describe('Task Detail View', () => {
		beforeEach(() => {
			TaskCommentFactory.truncate()
			LabelTaskFactory.truncate()
		})

		it('Shows all task details', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				index: 1,
				description: 'Lorem ipsum dolor sit amet.'
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
				done_at: new Date().toISOString()
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
				description: 'Lorem ipsum dolor sit amet.'
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .details.content.description .editor button')
				.click()
			cy.get('.task-view .details.content.description .editor .vue-easymde .EasyMDEContainer .CodeMirror-scroll')
				.type('{selectall}New Description')
			cy.get('[data-cy="saveEditor"]')
				.contains('Save')
				.click()

			cy.get('.task-view .details.content.description h3 span.is-small.has-text-success')
				.contains('Saved!')
				.should('exist')
		})

		it('Can add a new comment', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
			})
			cy.visit(`/tasks/${tasks[0].id}`)

			cy.get('.task-view .comments .media.comment .editor .vue-easymde .EasyMDEContainer .CodeMirror-scroll')
				.should('be.visible')
				.type('{selectall}New Comment')
			cy.get('.task-view .comments .media.comment .button:not([disabled])')
				.contains('Comment')
				.should('be.visible')
				.click()

			cy.get('.task-view .comments .media.comment .editor')
				.should('contain', 'New Comment')
			cy.get('.global-notification')
				.should('contain', 'Success')
		})

		it('Can move a task to another project', () => {
			const projects = ProjectFactory.create(2)
			BucketFactory.create(2, {
				project_id: '{increment}'
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
			cy.get('.modal-mask .modal-container .modal-content .header')
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
				bucket_id: buckets[0].id,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.truncate()
			
			cy.visit(`/projects/${projects[0].id}/kanban`)
			
			cy.get('.bucket .task')
				.contains(tasks[0].title)
				.click()
			
			addLabelToTaskAndVerify(labels[0].title)
			
			cy.get('.modal-content .close')
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
				bucket_id: buckets[0].id,
			})
			const labels = LabelFactory.create(1)
			LabelTaskFactory.truncate()

			cy.visit(`/projects/${projects[0].id}/kanban`)

			cy.get('.bucket .task')
				.contains(tasks[0].title)
				.click()

			uploadAttachmentAndVerify(tasks[0].id)

			cy.get('.modal-content .close')
				.click()

			cy.get('.bucket .task .footer .icon svg.fa-paperclip')
				.should('exist')
		})
		
		it('Can check items off a checklist', () => {
			const tasks = TaskFactory.create(1, {
				id: 1,
				description: `
This is a checklist:

* [ ] one item
* [ ] another item
* [ ] third item
* [ ] fourth item
* [x] and this one is already done
`,
			})
			cy.visit(`/tasks/${tasks[0].id}`)
			
			cy.get('.task-view .checklist-summary')
				.should('contain.text', '1 of 5 tasks')
			cy.get('.editor .content ul > li input[type=checkbox]')
				.eq(2)
				.click()
			
			cy.get('.editor .content ul > li input[type=checkbox]')
				.eq(2)
				.should('be.checked')
			cy.get('.editor .content input[type=checkbox]')
				.should('have.length', 5)
			cy.get('.task-view .checklist-summary')
				.should('contain.text', '2 of 5 tasks')
		})
	})
})
