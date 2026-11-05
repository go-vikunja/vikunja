import {test, expect} from '../../support/fixtures'
import {clickTaskAction, openTaskActions} from '../../support/taskDetail'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

test('task subscription survives reload and can be removed', async ({authenticatedPage: page, apiContext, userToken}) => {
	const [project] = await ProjectFactory.create(1)
	await createDefaultViews(project.id)
	const [task] = await TaskFactory.create(1, {project_id: project.id})
	const headers = {Authorization: `Bearer ${userToken}`}
	await page.goto(`/tasks/${task.id}`)
	await clickTaskAction(page, 'Subscribe')
	await expect((await openTaskActions(page)).getByRole('button', {name: 'Unsubscribe', exact: true})).toBeVisible()
	await page.reload()
	await clickTaskAction(page, 'Unsubscribe')
	await expect((await openTaskActions(page)).getByRole('button', {name: 'Subscribe', exact: true})).toBeVisible()
	await page.reload()
	await expect((await openTaskActions(page)).getByRole('button', {name: 'Subscribe', exact: true})).toBeVisible()
	const stored = await apiContext.get(`tasks/${task.id}`, {headers})
	expect(stored.ok()).toBeTruthy()
	expect((await stored.json()).subscription).toBeFalsy()
})
