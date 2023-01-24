import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareLists} from './prepareLists'

describe('Lists', () => {
	createFakeUserAndLogin()

	let lists
	prepareLists((newLists) => (lists = newLists))

	it('Should create a new list', () => {
		cy.visit('/')
		cy.get('.namespace-title .dropdown-trigger')
			.click()
		cy.get('.namespace-title .dropdown .dropdown-item')
			.contains('New list')
			.click()
		cy.url()
			.should('contain', '/lists/new/1')
		cy.get('.card-header-title')
			.contains('New list')
		cy.get('input.input')
			.type('New List')
		cy.get('.button')
			.contains('Create')
			.click()

		cy.get('.global-notification', { timeout: 1000 }) // Waiting until the request to create the new list is done
			.should('contain', 'Success')
		cy.url()
			.should('contain', '/lists/')
		cy.get('.list-title')
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

	it('Should rename the list in all places', () => {
		TaskFactory.create(5, {
			id: '{increment}',
			list_id: 1,
		})
		const newListName = 'New list name'

		cy.visit('/lists/1')
		cy.get('.list-title')
			.should('contain', 'First List')

		cy.get('.namespace-container .menu.namespaces-lists .menu-list li:first-child .dropdown .menu-list-dropdown-trigger')
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .menu-list li:first-child .dropdown .dropdown-content')
			.contains('Edit')
			.click()
		cy.get('#title')
			.type(`{selectall}${newListName}`)
		cy.get('footer.card-footer .button')
			.contains('Save')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.list-title')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
		cy.get('.namespace-container .menu.namespaces-lists .menu-list li:first-child')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
		cy.visit('/')
		cy.get('.card-content')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
	})

	it('Should remove a list', () => {
		cy.visit(`/lists/${lists[0].id}`)

		cy.get('.namespace-container .menu.namespaces-lists .menu-list li:first-child .dropdown .menu-list-dropdown-trigger')
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .menu-list li:first-child .dropdown .dropdown-content')
			.contains('Delete')
			.click()
		cy.url()
			.should('contain', '/settings/delete')
		cy.get('[data-cy="modalPrimary"]')
			.contains('Do it')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.namespace-container .menu.namespaces-lists .menu-list')
			.should('not.contain', lists[0].title)
		cy.location('pathname')
			.should('equal', '/')
	})
	
	it('Should archive a list', () => {
		cy.visit(`/lists/${lists[0].id}`)
		
		cy.get('.list-title-dropdown')
			.click()
		cy.get('.list-title-dropdown .dropdown-menu .dropdown-item')
			.contains('Archive')
			.click()
		cy.get('.modal-content')
			.should('contain.text', 'Archive this list')
		cy.get('.modal-content [data-cy=modalPrimary]')
			.click()
		
		cy.get('.namespace-container .menu.namespaces-lists .menu-list')
			.should('not.contain', lists[0].title)
		cy.get('main.app-content')
			.should('contain.text', 'This list is archived. It is not possible to create new or edit tasks for it.')
	})
})
