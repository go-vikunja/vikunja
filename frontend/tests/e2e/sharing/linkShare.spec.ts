import {test, expect} from '../../support/fixtures'
import {LinkShareFactory} from '../../factories/link_sharing'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {createProjects} from '../project/prepareProjects'

async function prepareLinkShare() {
	await UserFactory.create()
	const projects = await createProjects()
	const tasks = await TaskFactory.create(10, {
		project_id: projects[0].id,
	})
	const linkShares = await LinkShareFactory.create(1, {
		project_id: projects[0].id,
		permission: 0,
	})

	return {
		share: linkShares[0],
		project: projects[0],
		tasks,
	}
}

test.describe('Link shares', () => {
	test('Can view a link share', async ({page}) => {
		const {share, project, tasks} = await prepareLinkShare()

		await page.goto(`/share/${share.hash}/auth`)

		await expect(page.locator('h1.title')).toContainText(project.title)
		await expect(page.locator('input.input[placeholder="Add a task…"]')).not.toBeVisible()
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)

		await expect(page).toHaveURL(new RegExp(`/projects/${project.id}/1#share-auth-token=${share.hash}`))
	})

	test('Should work when directly viewing a project with share hash present', async ({page}) => {
		const {share, project, tasks} = await prepareLinkShare()

		await page.goto(`/projects/${project.id}/1#share-auth-token=${share.hash}`)

		await expect(page.locator('h1.title')).toContainText(project.title)
		await expect(page.locator('input.input[placeholder="Add a task…"]')).not.toBeVisible()
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)
	})

	test('Should work when directly viewing a task with share hash present', async ({page}) => {
		const {share, tasks} = await prepareLinkShare()

		await page.goto(`/tasks/${tasks[0].id}#share-auth-token=${share.hash}`)

		await expect(page.locator('h1.title.input')).toContainText(tasks[0].title)
	})
})
