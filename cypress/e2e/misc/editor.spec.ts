import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {NamespaceFactory} from '../../factories/namespace'
import {UserProjectFactory} from '../../factories/users_project'

describe('Editor', () => {
	createFakeUserAndLogin()
	
	beforeEach(() => {
		NamespaceFactory.create(1)
		ProjectFactory.create(1)
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