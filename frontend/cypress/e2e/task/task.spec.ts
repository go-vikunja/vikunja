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

	it.skip('Should be created new', () => {
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

	it.skip('Inserts new tasks at the top of the project', () => {
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

	it.skip('Marks a task as done', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/1')
		cy.get('.tasks .task .fancy-checkbox')
			.first()
			.click()
		cy.get('.global-notification')
			.should('contain', 'Success')
	})

	it.skip('Can add a task to favorites', () => {
		TaskFactory.create(1)

		cy.visit('/projects/1/1')
		cy.get('.tasks .task .favorite')
			.first()
			.click()
		cy.get('.menu-container')
			.should('contain', 'Favorites')
	})

	it.skip('Should show a task description icon if the task has a description', () => {
		cy.intercept('**/projects/1/views/*/tasks**').as('loadTasks')
		TaskFactory.create(1, {
			description: 'Lorem Ipsum',
		})

		cy.visit('/projects/1/1')
		cy.wait('@loadTasks')

		cy.get('.tasks .task .project-task-icon .fa-align-left')
			.should('exist')
	})

	it.skip('Should not show a task description icon if the task has an empty description', () => {
		cy.intercept('**/projects/1/views/*/tasks**').as('loadTasks')
		TaskFactory.create(1, {
			description: '',
		})

		cy.visit('/projects/1/1')
		cy.wait('@loadTasks')

		cy.get('.tasks .task .project-task-icon .fa-align-left')
			.should('not.exist')
	})

	it.skip('Should not show a task description icon if the task has a description containing only an empty p tag', () => {
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
		
	})
})
