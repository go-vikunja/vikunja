import {createFakeUserAndLogin} from '../../support/authenticateUser'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {createDefaultViews} from '../project/prepareProjects'

describe('Task comment pagination', () => {
	createFakeUserAndLogin()

	beforeEach(() => {
		ProjectFactory.create(1)
		createDefaultViews(1)
		TaskFactory.create(1, {id: 1})
		TaskCommentFactory.truncate()
	})

	it('shows pagination when more comments than configured page size', () => {
		cy.request('/api/v1/info').then((response) => {
			const pageSize = response.body.max_items_per_page
			TaskCommentFactory.create(pageSize + 10)
			cy.visit('/tasks/1')
			cy.get('.task-view .comments nav.pagination').should('exist')
		})
	})

	it('hides pagination when comments equal or fewer than configured page size', () => {
		cy.request('/api/v1/info').then((response) => {
			const pageSize = response.body.max_items_per_page
			TaskCommentFactory.create(Math.max(1, pageSize - 10))
			cy.visit('/tasks/1')
			cy.get('.task-view .comments nav.pagination').should('not.exist')
		})
	})
})
