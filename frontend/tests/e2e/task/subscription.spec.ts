import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

test('task subscription survives reload and can be removed', async ({authenticatedPage: page, apiContext, userToken}) => {
	const [project] = await ProjectFactory.create(1)
	await createDefaultViews(project.id)
	const [task] = await TaskFactory.create(1, {project_id: project.id})
	const headers = {Authorization: `Bearer ${userToken}`}
	await page.goto(`/tasks/${task.id}`)
	const actions = page.locator('.task-view .action-buttons')
	await actions.getByRole('button', {name: 'Subscribe', exact: true}).click()
	await expect(actions.getByRole('button', {name: 'Unsubscribe', exact: true})).toBeVisible()
	await page.reload()
	await actions.getByRole('button', {name: 'Unsubscribe', exact: true}).click()
	await expect(actions.getByRole('button', {name: 'Subscribe', exact: true})).toBeVisible()
	await page.reload()
	await expect(actions.getByRole('button', {name: 'Subscribe', exact: true})).toBeVisible()
	const stored = await apiContext.get(`tasks/${task.id}`, {headers})
	expect(stored.ok()).toBeTruthy()
	expect((await stored.json()).subscription).toBeFalsy()
})
