import {formatISO, format} from 'date-fns'

import {TaskFactory} from '../../factories/task'
import {ListFactory} from '../../factories/list'
import {UserListFactory} from '../../factories/users_list'
import {UserFactory} from '../../factories/user'
import {NamespaceFactory} from '../../factories/namespace'
import {BucketFactory} from '../../factories/bucket'

import '../../support/authenticateUser'

describe('Lists', () => {
	beforeEach(() => {
		UserFactory.create(1)
		NamespaceFactory.create(1)
		const lists = ListFactory.create(1, {
			title: 'First List'
		})
		TaskFactory.truncate()
	})

	it('Should create a new list', () => {
		cy.visit('/')
		cy.get('.namespace-title .dropdown-trigger')
			.click()
		cy.get('.namespace-title .dropdown .dropdown-item')
			.contains('New list')
			.click()
		cy.url()
			.should('contain', '/namespaces/1/list')
		cy.get('.card-header-title')
			.contains('Create a new list')
		cy.get('input.input')
			.type('New List')
		cy.get('.button')
			.contains('Create')
			.click()

		cy.wait(1000) // Waiting until the request to create the new list is done
		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.url()
			.should('contain', '/lists/')
		cy.get('.list-title h1')
			.should('contain', 'New List')
	})

	it('Should redirect to a specific list view after visited', () => {
		cy.visit('/lists/1/kanban')
		cy.url()
			.should('contain', '/lists/1/kanban')
		cy.visit('/lists/1')
		cy.url()
			.should('contain', '/lists/1/kanban')
	})

	describe('List View', () => {
		it('Should be an empty list', () => {
			cy.visit('/lists/1')
			cy.url()
				.should('contain', '/lists/1/list')
			cy.get('.list-title h1')
				.should('contain', 'First List')
			cy.get('.list-title .dropdown')
				.should('exist')
			cy.get('p')
				.contains('This list is currently empty.')
				.should('exist')
		})

		it('Should navigate to the task when the title is clicked', () => {
			const tasks = TaskFactory.create(5, {
				id: '{increment}',
				list_id: 1,
			})
			cy.visit('/lists/1/list')

			cy.get('.tasks .task .tasktext')
				.contains(tasks[0].title)
				.first()
				.click()

			cy.url()
				.should('contain', `/tasks/${tasks[0].id}`)
		})

		it('Should not see any elements for a list which is shared read only', () => {
			UserFactory.create(2)
			UserListFactory.create(1, {
				list_id: 2,
				user_id: 1,
				right: 0,
			})
			const lists = ListFactory.create(2, {
				owner_id: '{increment}',
				namespace_id: '{increment}',
			})
			cy.visit(`/lists/${lists[1].id}/`)

			cy.get('.list-title a.icon')
				.should('not.exist')
			cy.get('input.input[placeholder="Add a new task..."')
				.should('not.exist')
		})
	})

	describe('Table View', () => {
		it('Should show a table with tasks', () => {
			const tasks = TaskFactory.create(1)
			cy.visit('/lists/1/table')

			cy.get('.table-view table.table')
				.should('exist')
			cy.get('.table-view table.table')
				.should('contain', tasks[0].title)
		})

		it('Should have working column switches', () => {
			TaskFactory.create(1)
			cy.visit('/lists/1/table')

			cy.get('.table-view .filter-container .items .button')
				.contains('Columns')
				.click()
			cy.get('.table-view .filter-container .card .card-content .fancycheckbox .check')
				.contains('Priority')
				.click()
			cy.get('.table-view .filter-container .card .card-content .fancycheckbox .check')
				.contains('Done')
				.click()

			cy.get('.table-view table.table th')
				.contains('Priority')
				.should('exist')
			cy.get('.table-view table.table th')
				.contains('Done')
				.should('not.exist')
		})

		it('Should navigate to the task when the title is clicked', () => {
			const tasks = TaskFactory.create(5, {
				id: '{increment}',
				list_id: 1,
			})
			cy.visit('/lists/1/table')

			cy.get('.table-view table.table')
				.contains(tasks[0].title)
				.click()

			cy.url()
				.should('contain', `/tasks/${tasks[0].id}`)
		})
	})

	describe('Gantt View', () => {
		it('Hides tasks with no dates', () => {
			TaskFactory.create(1)
			cy.visit('/lists/1/gantt')

			cy.get('.gantt-chart-container .gantt-chart .tasks')
				.should('be.empty')
		})

		it('Shows tasks from the current and next month', () => {
			const now = new Date()
			cy.visit('/lists/1/gantt')

			cy.get('.gantt-chart-container .gantt-chart .months')
				.should('contain', format(now, 'MMMM'))
				.should('contain', format(now.setMonth(now.getMonth() + 1), 'MMMM'))
		})

		it('Shows tasks with dates', () => {
			const now = new Date()
			const tasks = TaskFactory.create(1, {
				start_date: formatISO(now),
				end_date: formatISO(now.setDate(now.getDate() + 4))
			})
			cy.visit('/lists/1/gantt')

			cy.get('.gantt-chart-container .gantt-chart .tasks')
				.should('not.be.empty')
			cy.get('.gantt-chart-container .gantt-chart .tasks')
				.should('contain', tasks[0].title)
		})

		it('Shows tasks with no dates after enabling them', () => {
			TaskFactory.create(1, {
				start_date: null,
				end_date: null,
			})
			cy.visit('/lists/1/gantt')

			cy.get('.gantt-chart-container .gantt-options .fancycheckbox')
				.contains('Show tasks which don\'t have dates set')
				.click()

			cy.get('.gantt-chart-container .gantt-chart .tasks')
				.should('not.be.empty')
			cy.get('.gantt-chart-container .gantt-chart .tasks .task.nodate')
				.should('exist')
		})

		it('Drags a task around', () => {
			const now = new Date()
			TaskFactory.create(1, {
				start_date: formatISO(now),
				end_date: formatISO(now.setDate(now.getDate() + 4))
			})
			cy.visit('/lists/1/gantt')

			cy.get('.gantt-chart-container .gantt-chart .tasks .task')
				.first()
				.trigger('mousedown', {which: 1})
				.trigger('mousemove', {clientX: 500, clientY: 0})
				.trigger('mouseup', {force: true})
		})
	})

	describe('Kanban', () => {
		let buckets

		beforeEach(() => {
			buckets = BucketFactory.create(2)
		})

		it('Shows all buckets with their tasks', () => {
			const data = TaskFactory.create(10, {
				list_id: 1,
				bucket_id: 1,
			})
			cy.visit('/lists/1/kanban')

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
			const data = TaskFactory.create(2, {
				list_id: 1,
				bucket_id: 1,
			})
			cy.visit('/lists/1/kanban')

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
			cy.visit('/lists/1/kanban')

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
			cy.visit('/lists/1/kanban')

			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-trigger')
				.first()
				.click()
			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item')
				.contains('Limit: Not set')
				.click()
			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item .field input.input')
				.first()
				.type(3)
			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item .field a.button.is-primary')
				.first()
				.click()

			cy.get('.kanban .bucket .bucket-header span.limit')
				.contains('0/3')
				.should('exist')
		})

		it('Can rename a bucket', () => {
			cy.visit('/lists/1/kanban')

			cy.get('.kanban .bucket .bucket-header .title')
				.first()
				.type('{selectall}New Bucket Title{enter}')
			cy.get('.kanban .bucket .bucket-header .title')
				.first()
				.should('contain', 'New Bucket Title')
		})

		it('Can delete a bucket', () => {
			cy.visit('/lists/1/kanban')

			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-trigger')
				.first()
				.click()
			cy.get('.kanban .bucket .bucket-header .dropdown.options .dropdown-menu .dropdown-item')
				.contains('Delete')
				.click()
			cy.get('.modal-mask .modal-container .modal-content .header')
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


		// The following test does not work. It seems like vue-smooth-dnd does not use either mousemove or dragstart
		// (not sure why this actually works at all?) and as I'm planning to swap that out for vuedraggable/sortable.js
		// anyway, I figured it wouldn't be worth the hassle right now.

//		it('Can drag tasks around', () => {
//			const tasks = TaskFactory.create(2, {
//				list_id: 1,
//				bucket_id: 1,
//			})
//			cy.visit('/lists/1/kanban')
//
//			cy.get('.kanban .bucket .tasks .task')
//				.contains(tasks[0].title)
//				.first()
//				.drag('.kanban .bucket:nth-child(2) .tasks .smooth-dnd-container.vertical')
//				.trigger('mousedown', {which: 1})
//				.trigger('mousemove', {clientX: 500, clientY: 0})
//				.trigger('mouseup', {force: true})
//		})

		it('Should navigate to the task when the task card is clicked', () => {
			const tasks = TaskFactory.create(5, {
				id: '{increment}',
				list_id: 1,
				bucket_id: 1,
			})
			cy.visit('/lists/1/kanban')

			cy.getAttached('.kanban .bucket .tasks .task')
				.contains(tasks[0].title)
				.should('be.visible')
				.click()

			cy.url()
				.should('contain', `/tasks/${tasks[0].id}`)
		})
	})
})
