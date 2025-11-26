import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {BucketFactory} from '../../factories/bucket'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'
import {ProjectViewFactory} from "../../factories/project_view";
import {TaskBucketFactory} from "../../factories/task_buckets";
import {
	createTasksWithPriorities,
	createTasksWithSearch,
} from '../../support/filterTestHelpers'

function createSingleTaskInBucket(count = 1, attrs = {}) {
	const projects = ProjectFactory.create(1)
	const views = ProjectViewFactory.create(1, {
		id: 1,
		project_id: projects[0].id,
		view_kind: 3,
		bucket_configuration_mode: 1,
	})
	const buckets = BucketFactory.create(2, {
		project_view_id: views[0].id,
	})
	const tasks = TaskFactory.create(count, {
		project_id: projects[0].id,
		...attrs,
	})
	TaskBucketFactory.create(1, {
		task_id: tasks[0].id,
		bucket_id: buckets[0].id,
		project_view_id: views[0].id,
	})
	return {
		task: tasks[0],
		view: views[0],
		project: projects[0],
	}
}

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

	it.skip('Can set a bucket limit', () => {
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
