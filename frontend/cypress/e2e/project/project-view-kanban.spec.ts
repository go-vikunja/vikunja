import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'
import {ProjectViewFactory} from "../../factories/project_view";
import {TaskBucketFactory} from "../../factories/task_buckets";

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
})
