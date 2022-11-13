import {LinkShareFactory} from '../../factories/link_sharing'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

describe('Link shares', () => {
	it('Can view a link share', () => {
		const projects = ProjectFactory.create(1)
		const tasks = TaskFactory.create(10, {
			project_id: projects[0].id
		})
		const linkShares = LinkShareFactory.create(1, {
			project_id: projects[0].id,
			right: 0,
		})

		cy.visit(`/share/${linkShares[0].hash}/auth`)

		cy.get('h1.title')
			.should('contain', projects[0].title)
		cy.get('input.input[placeholder="Add a new task..."')
			.should('not.exist')
		cy.get('.tasks')
			.should('contain', tasks[0].title)
	})
})
