import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {UserProjectFactory} from '../../factories/users_project'
import {BucketFactory} from '../../factories/bucket'

describe('Editor', () => {
	createFakeUserAndLogin()
	
	beforeEach(() => {
		ProjectFactory.create(1)
		BucketFactory.create(1)
		TaskFactory.truncate()
		UserProjectFactory.truncate()
	})

	it('Has a preview with checkable checkboxes', () => {
		const tasks = TaskFactory.create(1, {
			description: `# Test Heading
* Bullet 1
* Bullet 2

* [ ] Checklist
* [x] Checklist checked
`,
			bucket_id: 1,
		})

		cy.visit(`/tasks/${tasks[0].id}`)
		cy.get('input[type=checkbox][data-checkbox-num=0]')
			.click()

		cy.get('.task-view .details.content.description h3 span.is-small.has-text-success')
			.contains('Saved!')
			.should('exist')
		cy.get('.preview.content')
			.should('contain', 'Test Heading')
	})
})