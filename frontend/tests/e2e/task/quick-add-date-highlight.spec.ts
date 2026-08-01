import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {createDefaultViews} from '../project/prepareProjects'
import {login} from '../../support/authenticateUser'

test.describe('Quick add date highlighting', () => {
	test('Highlights parsed dates in the add task input', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await createDefaultViews(project.id)

		await login(page, apiContext, user)
		await page.goto(`/projects/${project.id}/1`)

		const input = page.locator('.input[placeholder="Add a task…"]')
		await input.fill('Buy milk tomorrow at 5pm')

		const fragments = page.locator('.task-add .date-fragment')
		await expect(fragments).toHaveCount(2)
		await expect(fragments.nth(0)).toHaveText('tomorrow')
		await expect(fragments.nth(1)).toHaveText('at 5pm')

		const fragment = fragments.nth(0)
		await expect(fragment).toHaveCSS('font-style', 'italic')

		// The highlight overlay must sit on top of the textarea text, not offset from it
		const inputBox = await input.boundingBox()
		const fragmentBox = await fragment.boundingBox()
		expect(inputBox).not.toBeNull()
		expect(fragmentBox).not.toBeNull()
		expect(fragmentBox!.x).toBeGreaterThanOrEqual(inputBox!.x)
		expect(fragmentBox!.x).toBeLessThan(inputBox!.x + inputBox!.width)
		expect(fragmentBox!.y).toBeGreaterThanOrEqual(inputBox!.y)
		expect(fragmentBox!.y).toBeLessThan(inputBox!.y + inputBox!.height)
	})

	test('Does not highlight quoted text', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await createDefaultViews(project.id)

		await login(page, apiContext, user)
		await page.goto(`/projects/${project.id}/1`)

		await page.locator('.input[placeholder="Add a task…"]').fill('"Buy milk tomorrow"')

		await expect(page.locator('.task-add .date-fragment')).toHaveCount(0)
	})
})
