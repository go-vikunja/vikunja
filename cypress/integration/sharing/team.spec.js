import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import '../../support/authenticateUser'

describe('Team', () => {
    it('Creates a new team', () => {
        TeamFactory.truncate()
        cy.visit('/teams')

        cy.get('a.button')
            .contains('New Team')
            .click()
		cy.url()
			.should('contain', '/teams/new')
		cy.get('h3')
			.contains('Create a new team')
		cy.get('input.input')
			.type('New Team')
		cy.get('button.is-success')
			.contains('Add')
			.click()

        cy.get('.fullpage')
            .should('not.exist')
		cy.url()
            .should('contain', '/edit')
        cy.get('.card-header .card-header-title')
            .first()
            .should('contain', 'Edit Team')
    })

    it('Shows all teams', () => {
        TeamMemberFactory.create(10, {
            team_id: '{increment}',
        })
        const teams = TeamFactory.create(10, {
            id: '{increment}',
        })

        cy.visit('/teams')

        cy.get('.teams.box')
            .should('not.be.empty')
        teams.forEach(t => {
            cy.get('.teams.box')
                .should('contain', t.name)
        })
    })

    it('Allows an admin to edit the team', () => {
        TeamMemberFactory.create(1, {
            team_id: 1,
            admin: true,
        })
        const teams = TeamFactory.create(1, {
            id: 1,
        })
        
        cy.visit('/teams/1/edit')
        cy.get('.card input.input')
            .first()
            .type('{selectall}New Team Name')

        cy.get('.card .button')
            .contains('Save')
            .click()
        
        cy.get('table.table td')
            .contains('Admin')
            .should('exist')
        cy.get('.global-notification')
			.should('contain', 'Success')
    })

    it('Does not allow a normal user to edit the team', () => {
        TeamMemberFactory.create(1, {
            team_id: 1,
            admin: false,
        })
        const teams = TeamFactory.create(1, {
            id: 1,
        })
        
        cy.visit('/teams/1/edit')
        cy.get('.card input.input')
            .should('not.exist')
        cy.get('table.table td')
            .contains('Member')
            .should('exist')
    })
})
