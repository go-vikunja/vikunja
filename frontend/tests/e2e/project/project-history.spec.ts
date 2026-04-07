import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {updateUserSettings} from '../../support/updateUserSettings'
import type {Page} from '@playwright/test'

async function visitProjectsToBuildHistory(page: Page, projects: any[]) {
	for (const project of projects) {
		const loadProjectPromise = page.waitForResponse(response =>
			response.url().includes(`/projects/${project.id}`) && response.request().method() === 'GET',
		)
		await page.goto(`/projects/${project.id}/${project.id}`)
		await loadProjectPromise
		await page.waitForFunction(
			(projectId) => {
				const history = JSON.parse(localStorage.getItem('projectHistory') || '[]')
				return history.some((h: any) => h.id === projectId)
			},
			project.id,
		)
	}
}

test.describe('Project History', () => {
	test('should show a project history on the home page', async ({authenticatedPage: page}) => {
		test.setTimeout(60000)
		const projects = await ProjectFactory.create(7)
		await ProjectViewFactory.truncate()
		for (const p of projects) {
			await ProjectViewFactory.create(1, {
				id: p.id,
				project_id: p.id,
			}, false)
		}

		const loadProjectArrayPromise = page.waitForResponse('**/api/v1/projects*')
		await page.goto('/')
		await loadProjectArrayPromise
		await expect(page.locator('body')).not.toContainText('Last viewed')

		for (let i = 0; i < projects.length; i++) {
			const loadProjectPromise = page.waitForResponse(response =>
				response.url().includes(`/projects/${projects[i].id}`) && response.request().method() === 'GET',
			)
			await page.goto(`/projects/${projects[i].id}/${projects[i].id}`)
			await loadProjectPromise
			// Wait for history to be saved to localStorage
			await page.waitForFunction(
				(projectId) => {
					const history = JSON.parse(localStorage.getItem('projectHistory') || '[]')
					return history.some((h: any) => h.id === projectId)
				},
				projects[i].id,
			)
		}

		await page.locator('nav.menu.top-menu a').filter({hasText: 'Overview'}).click()

		await expect(page.locator('body')).toContainText('Last viewed')
		await expect(page.locator('.project-grid')).not.toContainText(projects[0].title)
		await expect(page.locator('.project-grid')).toContainText(projects[1].title)
		await expect(page.locator('.project-grid')).toContainText(projects[2].title)
		await expect(page.locator('.project-grid')).toContainText(projects[3].title)
		await expect(page.locator('.project-grid')).toContainText(projects[4].title)
		await expect(page.locator('.project-grid')).toContainText(projects[5].title)
		await expect(page.locator('.project-grid')).toContainText(projects[6].title)
	})

	test('should hide the last viewed section when showLastViewed setting is disabled', async ({authenticatedPage: page, apiContext}) => {
		test.setTimeout(60000)
		const projects = await ProjectFactory.create(3)
		await ProjectViewFactory.truncate()
		for (const p of projects) {
			await ProjectViewFactory.create(1, {
				id: p.id,
				project_id: p.id,
			}, false)
		}

		// Visit projects to build up history
		await visitProjectsToBuildHistory(page, projects)

		// Go to overview and verify section is visible
		await page.goto('/')
		await expect(page.locator('body')).toContainText('Last viewed')

		// Disable the setting via API
		const token = await page.evaluate(() => localStorage.getItem('token'))
		await updateUserSettings(apiContext, token!, {
			frontendSettings: {
				showLastViewed: false,
			},
		})

		// Reload and verify section is hidden
		await page.reload()
		await page.waitForLoadState('networkidle')
		await expect(page.locator('body')).not.toContainText('Last viewed')
	})

	test('should show the last viewed section again when re-enabling showLastViewed', async ({authenticatedPage: page, apiContext}) => {
		test.setTimeout(60000)
		const projects = await ProjectFactory.create(2)
		await ProjectViewFactory.truncate()
		for (const p of projects) {
			await ProjectViewFactory.create(1, {
				id: p.id,
				project_id: p.id,
			}, false)
		}

		// Disable the setting first
		const token = await page.evaluate(() => localStorage.getItem('token'))
		await updateUserSettings(apiContext, token!, {
			frontendSettings: {
				showLastViewed: false,
			},
		})

		// Visit projects to build up history
		await visitProjectsToBuildHistory(page, projects)

		// Verify section is hidden
		await page.goto('/')
		await page.waitForLoadState('networkidle')
		await expect(page.locator('body')).not.toContainText('Last viewed')

		// Re-enable the setting
		await updateUserSettings(apiContext, token!, {
			frontendSettings: {
				showLastViewed: true,
			},
		})

		// Reload and verify section is visible again
		await page.reload()
		await page.waitForLoadState('networkidle')
		await expect(page.locator('body')).toContainText('Last viewed')
	})
})
