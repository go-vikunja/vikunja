import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'
import {ProjectViewFactory} from "../../factories/project_view";
import {TaskBucketFactory} from "../../factories/task_buckets";

function createSingleTaskInBucket(count = 1, attrs = {}) {
	const projects = ProjectFactory.create(1)
	const views = ProjectViewFactory.create(1, {
		id: 1,
		project_id: projects[0].id,
		view_kind: 3,
		bucket_configuration_mode: 1,
	})
	const buckets = BucketFactory.create(2, {
		project_view_id: views[0].id,
	})
	const tasks = TaskFactory.create(count, {
		project_id: projects[0].id,
		...attrs,
	})
	TaskBucketFactory.create(1, {
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

function createTaskWithBuckets(buckets, count = 1) {
	const data = TaskFactory.create(count, {
		project_id: 1,
	})
	TaskBucketFactory.truncate()
	data.forEach(t => TaskBucketFactory.create(1, {
		task_id: t.id,
		bucket_id: buckets[0].id,
		project_view_id: buckets[0].project_view_id,
	}, false))

	return data
}

describe('Project View Kanban', () => {
	createFakeUserAndLogin()
	prepareProjects()

	let buckets
	beforeEach(() => {
		buckets = BucketFactory.create(2, {
			project_view_id: 4,
		})
	})

	it('Shows all buckets with their tasks', () => {
		const data = createTaskWithBuckets(buckets, 10)
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .title')
			.contains(buckets[0].title)
			.should('exist')
		cy.get('.kanban .bucket .title')
			.contains(buckets[1].title)
			.should('exist')
		cy.get('.kanban .bucket')
			.first()
			.should('contain', data[0].title)
	})

	it('Can add a new task to a bucket', () => {
		createTaskWithBuckets(buckets, 2)
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket')
			.contains(buckets[0].title)
			.get('.bucket-footer .button')
			.contains('Add another task')
			.click()
		cy.get('.kanban .bucket')
			.contains(buckets[0].title)
			.get('.bucket-footer .field .control input.input')
			.type('New Task{enter}')

		cy.get('.kanban .bucket')
			.first()
			.should('contain', 'New Task')
	})

	it('Can create a new bucket', () => {
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket.new-bucket .button')
			.click()
		cy.get('.kanban .bucket.new-bucket input.input')
			.type('New Bucket{enter}')

		cy.wait(1000) // Wait for the request to finish
		cy.get('.kanban .bucket .title')
			.contains('New Bucket')
			.should('exist')
	})

	it('Can set a bucket limit', () => {
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-trigger')
			.first()
			.click()
		cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item')
			.contains('Limit: Not Set')
			.click()
		cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .field input.input')
			.first()
			.type('3')
		cy.get('[data-cy="setBucketLimit"]')
			.first()
			.click()

		cy.get('.kanban .bucket .bucket-header span.limit')
			.contains('0/3')
			.should('exist')
	})

	it('Can rename a bucket', () => {
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .bucket-header .title')
			.first()
			.type('{selectall}New Bucket Title{enter}')
		cy.get('.kanban .bucket .bucket-header .title')
			.first()
			.should('contain', 'New Bucket Title')
	})

	it('Can delete a bucket', () => {
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-trigger')
			.first()
			.click()
		cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item')
			.contains('Delete')
			.click()
		cy.get('.modal-mask .modal-container .modal-content .modal-header')
			.should('contain', 'Delete the bucket')
		cy.get('.modal-mask .modal-container .modal-content .actions .button')
			.contains('Do it!')
			.click()

		cy.get('.kanban .bucket .title')
			.contains(buckets[0].title)
			.should('not.exist')
		cy.get('.kanban .bucket .title')
			.contains(buckets[1].title)
			.should('exist')
	})

	it('Can drag tasks around', () => {
		const tasks = createTaskWithBuckets(buckets, 2)
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .tasks .task')
			.contains(tasks[0].title)
			.first()
			.drag('.kanban .bucket:nth-child(2) .tasks')

		cy.get('.kanban .bucket:nth-child(2) .tasks')
			.should('contain', tasks[0].title)
		cy.get('.kanban .bucket:nth-child(1) .tasks')
			.should('not.contain', tasks[0].title)
	})

	it('Should navigate to the task when the task card is clicked', () => {
		const tasks = createTaskWithBuckets(buckets, 5)
		cy.visit('/projects/1/4')

		cy.get('.kanban .bucket .tasks .task')
			.contains(tasks[0].title)
			.should('be.visible')
			.click()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`, {timeout: 1000})
	})

	it('Should remove a task from the kanban board when moving it to another project', () => {
		const projects = ProjectFactory.create(2)
		const views = ProjectViewFactory.create(2, {
			project_id: '{increment}',
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		BucketFactory.create(2)
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		TaskBucketFactory.create(5, {
			project_view_id: 1,
		})
		const task = tasks[0]
		cy.visit('/projects/1/'+views[0].id)

		cy.get('.kanban .bucket .tasks .task')
			.contains(task.title)
			.should('be.visible')
			.click()

		cy.get('.task-view .action-buttons .button', {timeout: 3000})
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

		cy.get('.global-notification', {timeout: 1000})
			.should('contain', 'Success')
		cy.go('back')
		cy.get('.kanban .bucket')
			.should('not.contain', task.title)
	})

	it('Shows a button to filter the kanban board', () => {
		cy.visit('/projects/1/4')

		cy.get('.project-kanban .filter-container .base-button')
			.should('exist')
	})

	it('Should remove a task from the board when deleting it', () => {
		const {task, view} = createSingleTaskInBucket(5)
		cy.visit(`/projects/1/${view.id}`)

		cy.get('.kanban .bucket .tasks .task')
			.contains(task.title)
			.should('be.visible')
			.click()
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

		cy.get('.kanban .bucket .tasks')
			.should('not.contain', task.title)
	})

	it('Should show a task description icon if the task has a description', () => {
		cy.intercept(Cypress.env('API_URL') + '/projects/1/views/*/tasks**').as('loadTasks')
		const {task, view} = createSingleTaskInBucket(1, {
			description: 'Lorem Ipsum',
		})

		cy.visit(`/projects/${task.project_id}/${view.id}`)
		cy.wait('@loadTasks')

		cy.get('.bucket .tasks .task .footer .icon svg')
			.should('exist')
	})

	it('Should not show a task description icon if the task has an empty description', () => {
		cy.intercept(Cypress.env('API_URL') + '/projects/1/views/*/tasks**').as('loadTasks')
		const {task, view} = createSingleTaskInBucket(1, {
			description: '',
		})

		cy.visit(`/projects/${task.project_id}/${view.id}`)
		cy.wait('@loadTasks')

		cy.get('.bucket .tasks .task .footer .icon svg')
			.should('not.exist')
	})

	it('Should not show a task description icon if the task has a description containing only an empty p tag', () => {
		cy.intercept(Cypress.env('API_URL') + '/projects/1/views/*/tasks**').as('loadTasks')
		const {task, view} = createSingleTaskInBucket(1, {
			description: '<p></p>',
		})

		cy.visit(`/projects/${task.project_id}/${view.id}`)
		cy.wait('@loadTasks')

		cy.get('.bucket .tasks .task .footer .icon svg')
			.should('not.exist')
	})
})
