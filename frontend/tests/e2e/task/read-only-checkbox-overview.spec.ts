import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {UserProjectFactory} from '../../factories/users_project'
import {UserFactory} from '../../factories/user'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Read-only checkbox on Overview', () => {
	test('Should disable checkboxes for tasks from read-only shared projects', async ({authenticatedPage: page, apiContext, currentUser}) => {
		// Create a second user who will own the shared project
		const [otherUser] = await UserFactory.create(1, {
			id: 2,
		}, false)

		// Create the own project (owned by test user, id=1)
		const [ownProject] = await ProjectFactory.create(1, {
			id: 1,
			title: 'Own Project',
			owner_id: currentUser.id,
		})
		const ownViews = await createDefaultViews(ownProject.id, 1)
		await BucketFactory.create(1, {
			project_view_id: ownViews[3].id,
		})

		// Create the shared project (owned by user 2)
		const [sharedProject] = await ProjectFactory.create(1, {
			id: 2,
			title: 'Shared Read-Only Project',
			owner_id: otherUser.id,
		}, false)
		const sharedViews = await createDefaultViews(sharedProject.id, 5, false)
		await BucketFactory.create(1, {
			id: 2,
			project_view_id: sharedViews[3].id,
		}, false)

		// Share the project read-only (permission=0) with the test user
		await UserProjectFactory.create(1, {
			id: 1,
			project_id: sharedProject.id,
			user_id: currentUser.id,
			permission: 0,
		})

		const now = new Date()
		const soon = new Date(now.getTime() + 24 * 60 * 60 * 1000) // tomorrow

		// Create a task in the own project
		await TaskFactory.create(1, {
			id: 1,
			title: 'Own Task',
			project_id: ownProject.id,
			created_by_id: currentUser.id,
			due_date: soon.toISOString(),
		})

		// Create a task in the shared read-only project
		await TaskFactory.create(1, {
			id: 2,
			title: 'Read Only Task',
			project_id: sharedProject.id,
			created_by_id: otherUser.id,
			due_date: soon.toISOString(),
		}, false)

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Wait for both tasks to appear on the overview
		const ownTaskRow = page.locator('.single-task', {hasText: 'Own Task'})
		const readOnlyTaskRow = page.locator('.single-task', {hasText: 'Read Only Task'})

		await expect(ownTaskRow).toBeVisible({timeout: 10000})
		await expect(readOnlyTaskRow).toBeVisible({timeout: 10000})

		// The checkbox for the own task should be enabled
		const ownCheckbox = ownTaskRow.locator('input[type="checkbox"]')
		await expect(ownCheckbox).toBeEnabled()

		// The checkbox for the read-only task should be disabled
		const readOnlyCheckbox = readOnlyTaskRow.locator('input[type="checkbox"]')
		await expect(readOnlyCheckbox).toBeDisabled()
	})
})
