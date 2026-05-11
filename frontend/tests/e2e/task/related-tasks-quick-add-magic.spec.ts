import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {createDefaultViews} from '../project/prepareProjects'
import {login} from '../../support/authenticateUser'

async function openRelatedTasksForm(page) {
	await page.locator('.task-view .action-buttons .button').filter({hasText: 'Add Relation'}).click()
	const input = page.locator('.task-relations .multiselect input').first()
	await expect(input).toBeVisible()
	return input
}

test.describe('Related tasks quick add magic', () => {
	test('Applies a label parsed via *prefix to the new related task', async ({authenticatedPage: page}) => {
		const project = (await ProjectFactory.create(1, {id: 1, title: 'Project A'}))[0]
		await createDefaultViews(project.id)
		const parent = (await TaskFactory.create(1, {id: 1, title: 'Parent task', project_id: project.id}, false))[0]

		await page.goto(`/tasks/${parent.id}`)
		const input = await openRelatedTasksForm(page)
		await input.fill('Subtask one *Urgent')
		await input.press('Enter')

		const relatedTaskLink = page.locator('.task-relations .related-tasks .task a').filter({hasText: 'Subtask one'})
		await expect(relatedTaskLink).toBeVisible({timeout: 10000})
		// Quick add magic strips the *Urgent prefix from the title
		await expect(relatedTaskLink).not.toContainText('*Urgent')

		await relatedTaskLink.click()
		await expect(page).toHaveURL(/\/tasks\/\d+/)
		await expect(page.locator('.task-view .details.labels-list .multiselect .input-wrapper span.tag').filter({hasText: 'Urgent'}))
			.toBeVisible({timeout: 10000})
	})

	test('Applies a priority parsed via !prefix to the new related task', async ({authenticatedPage: page}) => {
		const project = (await ProjectFactory.create(1, {id: 1, title: 'Project A'}))[0]
		await createDefaultViews(project.id)
		const parent = (await TaskFactory.create(1, {id: 1, title: 'Parent task', project_id: project.id}, false))[0]

		await page.goto(`/tasks/${parent.id}`)
		const input = await openRelatedTasksForm(page)
		await input.fill('Important work !4')
		await input.press('Enter')

		const relatedTaskLink = page.locator('.task-relations .related-tasks .task a').filter({hasText: 'Important work'})
		await expect(relatedTaskLink).toBeVisible({timeout: 10000})
		await expect(relatedTaskLink).not.toContainText('!4')

		await relatedTaskLink.click()
		// Priority 4 is "Urgent"
		await expect(page.locator('.task-view .columns.details select').first()).toHaveValue('4', {timeout: 10000})
	})

	test('Creates the related task in another project via +project prefix', async ({authenticatedPage: page}) => {
		const projectA = (await ProjectFactory.create(1, {id: 1, title: 'Source'}))[0]
		await createDefaultViews(projectA.id)
		const projectB = (await ProjectFactory.create(1, {id: 2, title: 'TargetProject'}, false))[0]
		await createDefaultViews(projectB.id, 5)
		const parent = (await TaskFactory.create(1, {id: 1, title: 'Parent task', project_id: projectA.id}, false))[0]

		await page.goto(`/tasks/${parent.id}`)
		const input = await openRelatedTasksForm(page)
		await input.fill('Cross task +TargetProject')
		await input.press('Enter')

		const relatedTaskRow = page.locator('.task-relations .related-tasks .task').filter({hasText: 'Cross task'})
		await expect(relatedTaskRow).toBeVisible({timeout: 10000})
		await expect(relatedTaskRow.locator('a')).not.toContainText('+TargetProject')
		// Cross-project marker shows the other project name
		await expect(relatedTaskRow.locator('.different-project')).toContainText('TargetProject')
	})

	test('Keeps the title literal when quick add magic is disabled', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({
				quickAddMagicMode: 'disabled',
			}),
		}))[0]
		const project = (await ProjectFactory.create(1, {id: 1, title: 'Project A', owner_id: user.id}))[0]
		await createDefaultViews(project.id)
		const parent = (await TaskFactory.create(1, {id: 1, title: 'Parent task', project_id: project.id, created_by_id: user.id}, false))[0]

		await login(page, apiContext, user)
		await page.goto(`/tasks/${parent.id}`)

		const input = await openRelatedTasksForm(page)
		await input.fill('Buy milk *Urgent')
		await input.press('Enter')

		// With magic disabled, the prefix stays in the title verbatim
		await expect(page.locator('.task-relations .related-tasks .task a').filter({hasText: 'Buy milk *Urgent'}))
			.toBeVisible({timeout: 10000})
	})
})
