import {UserFactory} from '../../factories/user'

const testAndAssertFailed = fixture => {
	cy.visit('/login')
	cy.get('input[id=username]').type(fixture.username)
	cy.get('input[id=password]').type(fixture.password)
	cy.get('button').contains('Login').click()

	cy.wait(5000) // It can take waaaayy too long to log the user in
	cy.url().should('include', '/')
	cy.get('div.notification.is-danger').contains('Wrong username or password.')
}

context('Login', () => {
	beforeEach(() => {
		UserFactory.create(1, {
			username: 'test',
		})
		cy.visit('/', {
			onBeforeLoad(win) {
				win.localStorage.removeItem('token')
			},
		})
	})

	it('Should log in with the right credentials', () => {
		const fixture = {
			username: 'test',
			password: '1234',
		}

		cy.visit('/login')
		cy.get('input[id=username]').type(fixture.username)
		cy.get('input[id=password]').type(fixture.password)
		cy.get('button').contains('Login').click()
		cy.url().should('include', '/')
		cy.get('h2').should('contain', `Hi ${fixture.username}!`)
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
})
