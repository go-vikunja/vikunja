import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {LicenseFactory} from '../../factories/license'
import {login} from '../../support/authenticateUser'

test.describe('Admin projects list', () => {
	test.beforeEach(async () => {
		await LicenseFactory.enable(['admin_panel'])
	})

	test.afterEach(async () => {
		await LicenseFactory.disable()
	})

	async function setup(page, apiContext) {
		const [admin] = await UserFactory.create(2, {
			username: (i: number) => i === 1 ? 'adminuser' : 'bobowner',
			is_admin: (i: number) => i === 1,
			default_project_id: (i: number) => i === 2 ? 3 : null,
		})
		await ProjectFactory.create(3, {
			title: (i: number) => ['Alpha project', 'Charlie project', 'Bob inbox'][i - 1],
			owner_id: (i: number) => i === 1 ? 1 : 2,
		})
		await login(page, apiContext, admin)
		await page.goto('/admin/projects')
		await expect(page.locator('.admin-projects tbody tr')).toHaveCount(3)
	}

	const titles = (page) => page.locator('.admin-projects tbody tr td:nth-child(2)')

	test('filters by title', async ({page, apiContext}) => {
		await setup(page, apiContext)

		await page.getByPlaceholder('Search projects…').fill('Alpha')

		await expect(titles(page)).toHaveText(['Alpha project'])
	})

	test('filters by owner', async ({page, apiContext}) => {
		await setup(page, apiContext)

		await page.getByPlaceholder('Filter by owner').pressSequentially('bobowner')
		await page.locator('.admin-projects__toolbar .search-results button', {hasText: 'bobowner'}).click()

		await expect(titles(page)).toHaveCount(2)
		await expect(titles(page)).not.toContainText(['Alpha project'])
	})

	test('hides inbox projects', async ({page, apiContext}) => {
		await setup(page, apiContext)

		await page.getByText('Hide inbox projects').click()

		await expect(titles(page)).toHaveCount(2)
		await expect(titles(page)).not.toContainText(['Bob inbox'])
	})

	test('sorts by a column', async ({page, apiContext}) => {
		await setup(page, apiContext)

		const sortByTitle = page.getByRole('button', {name: 'Sort by Title'})
		await sortByTitle.click()
		await expect(titles(page)).toHaveText(['Charlie project', 'Bob inbox', 'Alpha project'])

		await sortByTitle.click()
		await expect(titles(page)).toHaveText(['Alpha project', 'Bob inbox', 'Charlie project'])
	})
})
