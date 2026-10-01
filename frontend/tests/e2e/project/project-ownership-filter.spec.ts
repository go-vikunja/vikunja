import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {TeamProjectFactory} from '../../factories/team_project'
import {TaskFactory} from '../../factories/task'
import {SavedFilterFactory} from '../../factories/saved_filter'

test.beforeEach(async ({authenticatedPage, currentUser}) => {
	await UserFactory.create(1, {id: 2}, false)
	await ProjectFactory.create(5, {
		id: i => i,
		title: i => ['My project', 'Shared by me', 'Shared with me', 'Shared child', 'Archived shared'][i - 1],
		owner_id: i => i <= 2 ? currentUser.id : 2,
		parent_project_id: i => i === 4 ? 3 : 0,
		is_archived: i => i === 5,
	})
	await UserProjectFactory.create(3, {
		project_id: i => [2, 3, 5][i - 1],
		user_id: i => i === 1 ? 2 : currentUser.id,
	})
})

test('Filters owned and shared projects, including inherited access', async ({authenticatedPage: page}) => {
	await page.goto('/projects')
	const grid = page.locator('.project-grid')
	await expect(grid).toContainText('Shared child')
	await page.getByRole('button', {name: 'Personal', exact: true}).click()
	await expect(grid).toContainText('My project')
	await expect(grid).toContainText('Shared by me')
	await expect(grid.locator('.project-card')).toHaveCount(2)
	await page.getByRole('button', {name: 'Shared', exact: true}).click()
	await expect(grid).toContainText('Shared with me')
	await expect(grid).toContainText('Shared child')
	await expect(grid.locator('.project-card')).toHaveCount(2)
	await expect(page.getByRole('button', {name: 'Shared', exact: true})).toHaveAttribute('aria-pressed', 'true')
	await page.getByRole('button', {name: 'All', exact: true}).click()
	await expect(grid).toContainText('My project')
	await expect(grid).toContainText('Shared child')
	await expect(grid.locator('.project-card')).toHaveCount(4)
})

test('Combines ownership with archived visibility', async ({authenticatedPage: page}) => {
	await page.goto('/projects')
	await page.getByRole('button', {name: 'Shared', exact: true}).click()
	await expect(page.locator('.project-grid')).not.toContainText('Archived shared')
	await page.getByTestId('show-archived-check').click()
	await expect(page.locator('.project-grid')).toContainText('Archived shared')
	await page.getByRole('button', {name: 'Personal', exact: true}).click()
	await expect(page.locator('.project-grid .project-card')).toHaveCount(2)
})

test('Offers a way back when there are no shared projects', async ({authenticatedPage: page, currentUser}) => {
	await ProjectFactory.create(1, {owner_id: currentUser.id, title: 'Only mine'})
	await UserProjectFactory.truncate()
	await page.goto('/projects')
	await page.getByRole('button', {name: 'Shared', exact: true}).click()
	await expect(page.getByText('No projects shared with you.')).toBeVisible()
	await page.getByRole('button', {name: 'Show all projects', exact: true}).click()
	await expect(page.locator('.project-grid')).toContainText('Only mine')
})

test('Includes projects shared through a team', async ({authenticatedPage: page, currentUser}) => {
	await UserProjectFactory.truncate()
	await TeamFactory.create(1, {id: 1})
	await TeamMemberFactory.create(1, {team_id: 1, user_id: currentUser.id})
	await TeamProjectFactory.create(1, {team_id: 1, project_id: 3})
	await page.goto('/projects')
	await page.getByRole('button', {name: 'Shared', exact: true}).click()
	await expect(page.locator('.project-grid')).toContainText('Shared with me')
	await expect(page.locator('.project-grid')).toContainText('Shared child')
	await expect(page.locator('.project-grid .project-card')).toHaveCount(2)
})

test('Supports keyboard filtering on a narrow screen', async ({authenticatedPage: page}) => {
	await page.setViewportSize({width: 375, height: 812})
	await page.goto('/projects')
	const personal = page.getByRole('button', {name: 'Personal', exact: true})
	await personal.focus()
	await page.keyboard.press('Enter')
	await expect(personal).toHaveAttribute('aria-pressed', 'true')
	await expect(page.locator('.project-grid .project-card')).toHaveCount(2)
	await page.keyboard.press('Tab')
	await page.keyboard.press('Space')
	await expect(page.getByRole('button', {name: 'Shared', exact: true})).toHaveAttribute('aria-pressed', 'true')
	await expect(page.locator('.project-grid')).toContainText('Shared child')
	await page.keyboard.press('Tab')
	await page.keyboard.press('Space')
	await expect(page.getByRole('button', {name: 'Saved filters', exact: true})).toHaveAttribute('aria-pressed', 'true')
	expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
})

test('Shows saved filters separately from projects and favorites', async ({authenticatedPage: page, apiContext, userToken, currentUser}) => {
	await SavedFilterFactory.create(1, {title: 'My saved filter', owner_id: currentUser.id})
	const [task] = await TaskFactory.create(1, {project_id: 1})
	const response = await apiContext.put(`/api/v2/tasks/${task.id}`, {
		headers: {Authorization: `Bearer ${userToken}`},
		data: {title: task.title, is_favorite: true},
	})
	expect(response.ok()).toBe(true)
	await page.goto('/projects')
	const grid = page.locator('.project-grid')
	await expect(grid).toContainText('Favorites')
	await expect(grid).toContainText('My saved filter')
	for (const name of ['Personal', 'Shared']) {
		await page.getByRole('button', {name, exact: true}).click()
		await expect(grid).not.toContainText('Favorites')
		await expect(grid).not.toContainText('My saved filter')
	}
	await page.getByRole('button', {name: 'Saved filters', exact: true}).click()
	await expect(grid.locator('.project-card')).toHaveCount(1)
	await expect(grid).toContainText('My saved filter')
	await expect(grid).not.toContainText('Favorites')
	await page.getByTestId('show-archived-check').click()
	await expect(grid.locator('.project-card')).toHaveCount(1)
	await page.getByRole('button', {name: 'All', exact: true}).click()
	await expect(grid).toContainText('Favorites')
	await expect(grid).toContainText('My saved filter')
})

test('Shows an empty state when there are no saved filters', async ({authenticatedPage: page}) => {
	await page.goto('/projects')
	await page.getByRole('button', {name: 'Saved filters', exact: true}).click()
	await expect(page.getByText('No saved filters yet.')).toBeVisible()
	await page.getByRole('button', {name: 'Show all projects', exact: true}).click()
	await expect(page.locator('.project-grid')).toContainText('My project')
})
