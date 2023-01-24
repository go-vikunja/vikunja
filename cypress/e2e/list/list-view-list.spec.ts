import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {UserListFactory} from '../../factories/users_list'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {ListFactory} from '../../factories/list'
import {prepareLists} from './prepareLists'

describe('List View List', () => {
	createFakeUserAndLogin()
	prepareLists()

	it('Should be an empty list', () => {
		cy.visit('/lists/1')
		cy.url()
			.should('contain', '/lists/1/list')
		cy.get('.list-title')
			.should('contain', 'First List')
		cy.get('.list-title-dropdown')
			.should('exist')
		cy.get('p')
			.contains('This list is currently empty.')
			.should('exist')
	})
	
	it('Should create a new task', () => {
		const newTaskTitle = 'New task'
		
		cy.visit('/lists/1')
		cy.get('.task-add textarea')
			.type(newTaskTitle+'{enter}')
		cy.get('.tasks')
			.should('contain.text', newTaskTitle)
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

		cy.get('.list-title-wrapper .icon')
			.should('not.exist')
		cy.get('input.input[placeholder="Add a new task..."')
			.should('not.exist')
	})

	it('Should only show the color of a list in the navigation and not in the list view', () => {
		const lists = ListFactory.create(1, {
			hex_color: '00db60',
		})
		TaskFactory.create(10, {
			list_id: lists[0].id,
		})
		cy.visit(`/lists/${lists[0].id}/`)

		cy.get('.menu-list li .list-menu-link .color-bubble')
			.should('have.css', 'background-color', 'rgb(0, 219, 96)')
		cy.get('.tasks .color-bubble')
			.should('not.exist')
	})

	it('Should paginate for > 50 tasks', () => {
		const tasks = TaskFactory.create(100, {
			id: '{increment}',
			title: i => `task${i}`,
			list_id: 1,
		})
		cy.visit('/lists/1/list')

		cy.get('.tasks')
			.should('contain', tasks[1].title)
		cy.get('.tasks')
			.should('not.contain', tasks[99].title)

		cy.get('.card-content .pagination .pagination-link')
			.contains('2')
			.click()

		cy.url()
			.should('contain', '?page=2')
		cy.get('.tasks')
			.should('contain', tasks[99].title)
		cy.get('.tasks')
			.should('not.contain', tasks[1].title)
	})
})