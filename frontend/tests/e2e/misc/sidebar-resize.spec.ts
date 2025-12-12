import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews} from '../project/prepareProjects'

async function seedTasks(userId: number, numberOfTasks = 5, startDueDate = new Date()) {
	const project = (await ProjectFactory.create(1, {owner_id: userId}))[0]
	const views = await createDefaultViews(project.id)
	await BucketFactory.create(1, {
		project_view_id: views[3].id,
	})
	const tasks = []
	let dueDate = startDueDate
	for (let i = 0; i < numberOfTasks; i++) {
		const now = new Date()
		dueDate = new Date(new Date(dueDate).setDate(dueDate.getDate() + 2))
		tasks.push({
			id: i + 1,
			project_id: project.id,
			done: false,
			created_by_id: userId,
			title: 'Test Task ' + i,
			index: i + 1,
			due_date: dueDate.toISOString(),
			created: now.toISOString(),
			updated: now.toISOString(),
		})
	}
	await TaskFactory.seed(TaskFactory.table, tasks)
	return {tasks, project}
}

test.describe('Sidebar Resize', () => {
	test('should not reload tasks when resizing the sidebar', async ({authenticatedPage: page, currentUser}) => {
		console.log('DEBUG: currentUser.id =', currentUser.id, 'username =', currentUser.username)
		await page.setViewportSize({width: 1280, height: 720})
		const {project, tasks} = await seedTasks(currentUser.id, 5)
		console.log('DEBUG: created project.id =', project.id, 'owner_id =', project.owner_id, 'task count =', tasks.length)

		// Debug: Intercept API responses before navigating
		let projectsResponse: any = null
		let tasksResponse: any = null
		page.on('response', async response => {
			if (response.url().includes('/projects') && response.request().method() === 'GET') {
				try {
					projectsResponse = await response.json()
					console.log('DEBUG: /projects response:', JSON.stringify(projectsResponse))
				} catch (e) {
					console.log('DEBUG: /projects response not JSON:', response.status())
				}
			}
			if (response.url().includes('/tasks/all') && response.request().method() === 'GET') {
				try {
					tasksResponse = await response.json()
					console.log('DEBUG: /tasks/all response count:', Array.isArray(tasksResponse) ? tasksResponse.length : 'not array')
				} catch (e) {
					console.log('DEBUG: /tasks/all response not JSON:', response.status())
				}
			}
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Debug: Check localStorage
		const apiUrl = await page.evaluate(() => window.localStorage.getItem('API_URL'))
		const token = await page.evaluate(() => window.localStorage.getItem('token'))
		console.log('DEBUG: localStorage API_URL =', apiUrl)
		console.log('DEBUG: localStorage token =', token ? 'SET' : 'NOT SET')

		// Debug: Check window.API_URL
		const windowApiUrl = await page.evaluate(() => window.API_URL)
		console.log('DEBUG: window.API_URL =', windowApiUrl)

		// Debug: check elements
		const showTasksExists = await page.locator('[data-cy="showTasks"]').count()
		console.log('DEBUG: showTasks count =', showTasksExists)

		// Debug: Wait for potential async loading and check store state
		await page.waitForTimeout(2000)
		const storeDebug = await page.evaluate(() => {
			// @ts-ignore - accessing Pinia store from window
			const projectStore = window.__pinia?.state?.value?.project
			const hasProjects = projectStore?.projects ? Object.keys(projectStore.projects).length > 0 : false
			const projectCount = projectStore?.projects ? Object.keys(projectStore.projects).length : 0
			return {
				hasProjects,
				projectCount,
				projectIds: projectStore?.projects ? Object.keys(projectStore.projects) : [],
			}
		})
		console.log('DEBUG: projectStore state:', JSON.stringify(storeDebug))

		// Debug: Check card and task elements
		const cardCount = await page.locator('[data-cy="showTasks"] .card').count()
		const taskCount = await page.locator('[data-cy="showTasks"] .card .task').count()
		console.log('DEBUG: card count =', cardCount, 'task count =', taskCount)

		await expect(page.locator('[data-cy="showTasks"] .card .task').first()).toBeVisible({timeout: 10000})

		let taskApiCalls = 0
		page.on('request', request => {
			if (request.url().includes('/tasks') && request.method() === 'GET') {
				taskApiCalls++
			}
		})

		await page.waitForTimeout(500)
		taskApiCalls = 0

		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).toBeAttached()

		const sidebar = page.locator('.menu-container')
		const sidebarBox = await sidebar.boundingBox()
		expect(sidebarBox).not.toBeNull()

		const startX = sidebarBox!.x + sidebarBox!.width - 2
		const startY = sidebarBox!.y + sidebarBox!.height / 2

		await page.mouse.move(startX, startY)
		await page.mouse.down()
		await page.mouse.move(startX + 100, startY, {steps: 10})
		await page.mouse.up()

		await page.waitForTimeout(1500)

		expect(taskApiCalls).toBe(0)
	})

	test('should persist sidebar width after resize', async ({authenticatedPage: page, currentUser}) => {
		await page.setViewportSize({width: 1280, height: 720})
		await seedTasks(currentUser.id, 1)

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		const sidebar = page.locator('.menu-container')
		const initialBox = await sidebar.boundingBox()
		expect(initialBox).not.toBeNull()
		const initialWidth = initialBox!.width

		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).toBeAttached()

		const startX = initialBox!.x + initialBox!.width - 2
		const startY = initialBox!.y + initialBox!.height / 2

		await page.mouse.move(startX, startY)
		await page.mouse.down()
		await page.mouse.move(startX + 50, startY, {steps: 10})
		await page.mouse.up()

		await page.waitForTimeout(1000)

		const newBox = await sidebar.boundingBox()
		expect(newBox).not.toBeNull()
		expect(newBox!.width).toBeGreaterThan(initialWidth)

		await page.reload()
		await page.waitForLoadState('networkidle')

		const reloadedBox = await sidebar.boundingBox()
		expect(reloadedBox).not.toBeNull()
		expect(Math.abs(reloadedBox!.width - newBox!.width)).toBeLessThan(5)
	})

	test('should not show resize handle on mobile', async ({authenticatedPage: page, currentUser}) => {
		await seedTasks(currentUser.id, 1)
		await page.setViewportSize({width: 375, height: 667})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).not.toBeAttached()
	})
})
