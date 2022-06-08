import {TaskFactory} from '../../factories/task'
import {prepareLists} from './prepareLists'

import '../../support/authenticateUser'

describe('Lists', () => {
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

	it('Should rename the list in all places', () => {
		TaskFactory.create(5, {
			id: '{increment}',
			list_id: 1,
		})
		const newListName = 'New list name'

		cy.visit('/lists/1')
		cy.get('.list-title h1')
			.should('contain', 'First List')

		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list li:first-child .dropdown .dropdown-trigger')
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list li:first-child .dropdown .dropdown-content')
			.contains('Edit')
			.click()
		cy.get('#title')
			.type(`{selectall}${newListName}`)
		cy.get('footer.modal-card-foot .button')
			.contains('Save')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.list-title h1')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list li:first-child')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
		cy.visit('/')
		cy.get('.card-content')
			.should('contain', newListName)
			.should('not.contain', lists[0].title)
	})

	it('Should remove a list', () => {
		cy.visit(`/lists/${lists[0].id}`)

		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list li:first-child .dropdown .dropdown-trigger')
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list li:first-child .dropdown .dropdown-content')
			.contains('Delete')
			.click()
		cy.url()
			.should('contain', '/settings/delete')
		cy.get('[data-cy="modalPrimary"]')
			.contains('Do it')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.namespace-container .menu.namespaces-lists .more-container .menu-list')
			.should('not.contain', lists[0].title)
		cy.location('pathname')
			.should('equal', '/')
	})
})
