import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'

const testAndAssertFailed = fixture => {
	cy.intercept(Cypress.env('API_URL') + '/login*').as('login')
	
	cy.visit('/login')
	cy.get('input[id=username]').type(fixture.username)
	cy.get('input[id=password]').type(fixture.password)
	cy.get('.button').contains('Login').click()

	cy.wait('@login')
	cy.url().should('include', '/')
	cy.get('div.message.danger').contains('Wrong username or password.')
}

const credentials = {
	username: 'test',
	password: '1234',
}

function login() {
	cy.get('input[id=username]').type(credentials.username)
	cy.get('input[id=password]').type(credentials.password)
	cy.get('.button').contains('Login').click()
	cy.url().should('include', '/')
}

context('Login', () => {
	beforeEach(() => {
		UserFactory.create(1, {username: credentials.username})
	})

	it('Should log in with the right credentials', () => {
		cy.visit('/login')
		login()
		cy.clock(1625656161057) // 13:00
		cy.get('h2').should('contain', `Hi ${credentials.username}!`)
	})

	it('Should fail with a bad password', () => {
		const fixture = {
			username: 'test',
			password: '123456',
		}

		testAndAssertFailed(fixture)
	})

	it('Should fail with a bad username', () => {
		const fixture = {
			username: 'loremipsum',
			password: '1234',
		}

		testAndAssertFailed(fixture)
	})
	
	it('Should redirect to /login when no user is logged in', () => {
		cy.visit('/')
		cy.url().should('include', '/login')
	})
	
	it('Should redirect to the previous route after logging in', () => {
		const projects = ProjectFactory.create(1)
		cy.visit(`/projects/${projects[0].id}/1`)

		cy.url().should('include', '/login')
		
		login()

		cy.url().should('include', `/projects/${projects[0].id}/1`)
	})
})
