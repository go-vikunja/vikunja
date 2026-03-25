import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {TaskFactory} from '../../factories/task'
import {login} from '../../support/authenticateUser'
import {updateUserSettings} from '../../support/updateUserSettings'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Default Landing Page', () => {
	test('shows overview page with default settings when no last visited page exists', async ({authenticatedPage: page}) => {
		await page.goto('/')
		await page.waitForLoadState('networkidle')
		await expect(page).toHaveURL('/')
		await expect(page.locator('.home.app-content')).toBeVisible()
	})

	test('redirects to upcoming when set as default page', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({defaultPage: 'upcoming'}),
		}))[0]
		await ProjectFactory.create(1, {owner_id: user.id})
		await login(page, apiContext, user)

		await page.goto('/')
		await page.waitForURL('**/tasks/range**')
	})

	test('redirects to default project when set as default page', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({defaultPage: 'defaultProject'}),
		}))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await createDefaultViews(project.id)

		const {token} = await login(page, apiContext, user)

		await updateUserSettings(apiContext, token, {
			default_project_id: project.id,
			overdue_tasks_reminders_time: '9:00',
		})

		await page.goto('/')
		await page.waitForURL(`**/projects/${project.id}/**`)
	})

	test('falls back to overview when default project does not exist', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({defaultPage: 'defaultProject'}),
		}))[0]
		await ProjectFactory.create(1, {owner_id: user.id})

		const {token} = await login(page, apiContext, user)

		await updateUserSettings(apiContext, token, {
			default_project_id: 999999,
			overdue_tasks_reminders_time: '9:00',
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')
		await expect(page).toHaveURL('/')
	})

	test('redirects to last visited page when set as default page', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({defaultPage: 'lastVisited'}),
		}))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		const views = await createDefaultViews(project.id)
		await TaskFactory.create(1, {project_id: project.id})

		await login(page, apiContext, user)

		await page.goto(`/projects/${project.id}/${views[0].id}`)
		await page.waitForLoadState('networkidle')

		await page.goto('/')
		await page.waitForURL(`**/projects/${project.id}/${views[0].id}`)
	})

	test('does not redirect on in-app navigation to home', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({defaultPage: 'upcoming'}),
		}))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await createDefaultViews(project.id)
		await TaskFactory.create(1, {project_id: project.id})

		await login(page, apiContext, user)

		await page.goto(`/projects/${project.id}/1`)
		await page.waitForLoadState('networkidle')

		await page.locator('.logo-link').click()
		await page.waitForLoadState('networkidle')
		await expect(page).toHaveURL('/')
	})
})
