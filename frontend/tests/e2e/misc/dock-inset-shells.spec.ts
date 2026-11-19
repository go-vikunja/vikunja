import type {Page} from '@playwright/test'
import {readFile} from 'node:fs/promises'
import {test, expect} from '../../support/fixtures'
import {login, setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD_HASH} from '../../support/constants'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {LinkShareFactory} from '../../factories/link_sharing'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {createDefaultViews} from '../project/prepareProjects'

const DOCK_TOOLBAR = 40
const WINDOW = {
	width: 1280,
	height: 720,
}
const PHONE_WINDOW = {
	width: 390,
	height: 844,
}
const NO_AUTH_ROOT = '.no-auth-wrapper'
const SHARE_ROOT = '.link-share-container'
const SHARE_VIEWS = [
	{
		name: 'list',
		index: 0,
		selector: '.tasks',
	},
	{
		name: 'table',
		index: 2,
		selector: '.project-table table',
	},
	{
		name: 'kanban',
		index: 3,
		selector: '.kanban .bucket',
	},
	{
		name: 'gantt',
		index: 1,
		selector: '.gantt-rows',
	},
]

async function behaveLikeDockApp(page: Page) {
	await page.addInitScript(() => {
		Object.defineProperty(navigator, 'standalone', {get: () => true})
	})
}

function top(page: Page, selector: string) {
	return page.locator(selector).first().evaluate(element => element.getBoundingClientRect().top)
}

interface ShellCase {
	url: string
	root: string
	selector: string
	assertIdentity: (page: Page) => Promise<void>
	verify?: (page: Page, dock: boolean) => Promise<void>
}

async function checkShellInset(page: Page, shell: ShellCase) {
	let normalTop = 0
	for (const dock of [false, true]) {
		if (dock) {
			// Re-authenticate shares on a fresh app load, including logoVisible=false.
			await page.evaluate(() => localStorage.clear())
			await behaveLikeDockApp(page)
		}
		await page.goto(shell.url)
		await shell.assertIdentity(page)
		await expect(page.locator(shell.selector).first()).toBeVisible()
		await page.evaluate(async () => {
			await document.fonts.ready
			window.scrollTo(0, 0)
		})
		// Pin the normal-tab fallback as well as the shell's own Dock declaration.
		await expect(page.locator(shell.root)).toHaveCSS('padding-block-start', dock ? '40px' : '0px')
		const position = await top(page, shell.selector)
		if (dock) {
			expect(position).toBeGreaterThanOrEqual(DOCK_TOOLBAR)
			expect(position).toBeCloseTo(normalTop + DOCK_TOOLBAR, 5)
		} else {
			normalTop = position
		}
		await shell.verify?.(page, dock)
	}
}

const loggedOutPages = [
	{
		name: 'login',
		url: '/login',
		field: '#loginform #username',
	},
	{
		name: 'register',
		url: '/register',
		field: '#registerform #email',
	},
	{
		name: 'password-reset request',
		url: '/get-password-reset',
		field: 'form #email',
	},
	{
		name: 'password reset',
		url: '/password-reset?userPasswordReset=dock-layout-token',
		field: '#form input[type="password"]',
	},
]

function loggedOutShell(route: typeof loggedOutPages[number]): ShellCase {
	return {
		url: route.url,
		root: NO_AUTH_ROOT,
		selector: `${NO_AUTH_ROOT} .logo`,
		async assertIdentity(page) {
			await expect(page).toHaveURL(route.url)
			await expect(page.locator(route.field)).toBeVisible()
			if (route.name === 'password-reset request') {
				await expect(page.locator('form button[type="submit"]')).toBeVisible()
			}
		},
	}
}

async function prepareShare() {
	// Factory.create loses the seeded row type; these fixtures use numeric IDs.
	const [owner] = await UserFactory.create(1) as {id: number}[]
	const [project] = await ProjectFactory.create(1, {
		title: 'Dock share',
		owner_id: owner.id,
	}) as {id: number}[]
	const views = await createDefaultViews(project.id) as {id: number}[]
	const start = new Date()
	start.setHours(12, 0, 0, 0)
	const end = new Date(start)
	end.setDate(end.getDate() + 1)
	const [task] = await TaskFactory.create(1, {
		title: 'Dock task',
		project_id: project.id,
		start_date: start.toISOString(),
		end_date: end.toISOString(),
	}) as {id: number}[]
	const [share] = await LinkShareFactory.create(1, {
		project_id: project.id,
		shared_by_id: owner.id,
	}) as {hash: string}[]
	return {
		owner,
		project,
		views,
		share,
		task,
	}
}

function shareShell(projectId: number, viewId: number, hash: string, viewSelector = '.tasks'): ShellCase {
	return {
		url: `/share/${hash}/auth?view=${viewId}`,
		root: SHARE_ROOT,
		selector: `${SHARE_ROOT} .logo:visible, ${SHARE_ROOT} .title:visible`,
		async assertIdentity(page) {
			await expect(page).toHaveURL(`/projects/${projectId}/${viewId}#share-auth-token=${hash}`)
			await expect(page.locator(`${SHARE_ROOT} h1.title`)).toHaveText('Dock share')
			await expect(page.locator(viewSelector)).toContainText('Dock task')
		},
	}
}

function sharePasswordShell(hash: string): ShellCase {
	const url = `/share/${hash}/auth`
	return {
		url,
		root: NO_AUTH_ROOT,
		selector: `${NO_AUTH_ROOT} .logo`,
		async assertIdentity(page) {
			await expect(page).toHaveURL(url)
			await expect(page.locator('#linkSharePassword')).toBeVisible()
			await expect(page.getByRole('button', {
				name: 'Login',
				exact: true,
			})).toBeVisible()
		},
	}
}

test.describe('Page shells below the Safari Dock toolbar', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
		await page.setViewportSize(WINDOW)
	})

	for (const route of loggedOutPages) {
		test(`${route.name} keeps normal padding and moves its logo by 40px in a Dock app`, async ({page}) => {
			await checkShellInset(page, loggedOutShell(route))
		})
	}

	for (const viewport of [
		{
			width: WINDOW.width,
			paddingBlockEnd: '240px',
		},
		{
			width: 1536,
			paddingBlockEnd: '0px',
		},
	]) {
		test(`login keeps ${viewport.paddingBlockEnd} bottom padding at ${viewport.width}px in both modes`, async ({page}) => {
			await page.setViewportSize({
				width: viewport.width,
				height: WINDOW.height,
			})
			await checkShellInset(page, {
				...loggedOutShell(loggedOutPages[0]),
				async verify(page) {
					await expect(page.locator(NO_AUTH_ROOT)).toHaveCSS('padding-block-end', viewport.paddingBlockEnd)
				},
			})
		})
	}

	test('a list share keeps normal padding and moves its first visible logo or title by 40px', async ({page}) => {
		const {project, views, share} = await prepareShare()
		await checkShellInset(page, shareShell(project.id, views[0].id, share.hash))
	})

	for (const view of SHARE_VIEWS.filter(view => view.name !== 'list')) {
		test(`a ${view.name} share moves its heading by 40px in a Dock app`, async ({page}) => {
			const {project, views, share, task} = await prepareShare()
			if (view.name === 'kanban') {
				const [bucket] = await BucketFactory.create(1, {
					title: 'Dock bucket',
					project_view_id: views[view.index].id,
				}) as {id: number}[]
				await TaskBucketFactory.create(1, {
					task_id: task.id,
					bucket_id: bucket.id,
					project_view_id: views[view.index].id,
				})
			}
			await checkShellInset(page, {
				...shareShell(project.id, views[view.index].id, share.hash, view.selector),
				selector: `${SHARE_ROOT} h1.title`,
			})
		})
	}

	test('a share with its logo hidden still moves the project title below the toolbar', async ({page}) => {
		const {project, views, share} = await prepareShare()
		const shell = shareShell(project.id, views[0].id, share.hash)
		await checkShellInset(page, {
			...shell,
			url: `${shell.url}&logoVisible=false`,
			selector: `${SHARE_ROOT} h1.title`,
			async verify(page) {
				await expect(page.locator(`${SHARE_ROOT} .logo`)).toHaveCount(0)
			},
		})
	})

	test('a share background covers the toolbar area while its content moves by 40px', async ({page, apiContext}) => {
		const {owner, project, views, share} = await prepareShare()
		const {token} = await login(null, apiContext, owner)
		// Use the real upload API, as project.spec.ts does through the settings UI.
		const upload = await apiContext.put(`/api/v2/projects/${project.id}/backgrounds/upload`, {
			headers: {Authorization: `Bearer ${token}`},
			multipart: {
				background: {
					name: 'image-blue.png',
					mimeType: 'image/png',
					buffer: await readFile('tests/fixtures/image-blue.png'),
				},
			},
		})
		expect(upload.ok()).toBe(true)
		await checkShellInset(page, {
			...shareShell(project.id, views[0].id, share.hash),
			async verify(page, dock) {
				const root = page.locator(SHARE_ROOT)
				await expect(root).toHaveClass(/has-background/)
				await expect(root).toHaveCSS('background-image', /url\(["']?blob:/)
				await expect(root).toHaveCSS('background-attachment', 'fixed')
				await expect(root).toHaveCSS('background-size', 'cover')
				await expect(root).toHaveCSS('background-clip', 'border-box')
				await expect(root).toHaveCSS('min-height', `${WINDOW.height}px`)
				if (dock) {
					expect(await top(page, SHARE_ROOT)).toBe(0)
				}
			},
		})
	})

	test('the password-protected share auth page moves its logo by 40px', async ({page}) => {
		const {project} = await prepareShare()
		const [share] = await LinkShareFactory.create(1, {
			project_id: project.id,
			sharing_type: 2,
			password: TEST_PASSWORD_HASH,
		}) as {hash: string}[]
		await checkShellInset(page, sharePasswordShell(share.hash))
	})

	test('a short login window keeps the logo clear and can scroll to the submit button', async ({page}) => {
		await page.setViewportSize({
			width: 1280,
			height: 400,
		})
		await checkShellInset(page, {
			...loggedOutShell(loggedOutPages[0]),
			async verify(page) {
				const submit = page.locator('#loginform').getByRole('button', {
					name: 'Login',
					exact: true,
				})
				await submit.scrollIntoViewIfNeeded()
				await expect(submit).toBeInViewport({ratio: 1})
				expect(await page.evaluate(() => document.documentElement.scrollHeight)).toBeGreaterThan(400)
			},
		})
	})

	for (const viewport of [WINDOW, PHONE_WINDOW]) {
		for (const name of [...loggedOutPages.map(route => route.name), ...SHARE_VIEWS.map(view => `${view.name} share`), 'hidden-logo share', 'share password']) {
			test(`${name} has no document overflow at ${viewport.width}x${viewport.height} in Dock mode`, async ({page}) => {
				await page.setViewportSize(viewport)
				await behaveLikeDockApp(page)
				const route = loggedOutPages.find(route => route.name === name)
				let shell = loggedOutShell(route ?? loggedOutPages[0])
				if (!route) {
					const {project, views, share, task} = await prepareShare()
					if (name === 'share password') {
						const [protectedShare] = await LinkShareFactory.create(1, {
							project_id: project.id,
							sharing_type: 2,
							password: TEST_PASSWORD_HASH,
						}) as {hash: string}[]
						shell = sharePasswordShell(protectedShare.hash)
					} else {
						const view = SHARE_VIEWS.find(view => name === `${view.name} share`) ?? SHARE_VIEWS[0]
						if (view.name === 'kanban') {
							const [bucket] = await BucketFactory.create(1, {
								title: 'Dock bucket',
								project_view_id: views[view.index].id,
							}) as {id: number}[]
							await TaskBucketFactory.create(1, {
								task_id: task.id,
								bucket_id: bucket.id,
								project_view_id: views[view.index].id,
							})
						}
						shell = shareShell(project.id, views[view.index].id, share.hash, view.selector)
						if (name === 'hidden-logo share') shell.url += '&logoVisible=false'
					}
				}
				await page.goto(shell.url)
				await shell.assertIdentity(page)
				if (name === 'hidden-logo share') {
					await expect(page.locator(`${SHARE_ROOT} .logo`)).toHaveCount(0)
				}
				await expect(page.locator(shell.root)).toHaveCSS('padding-block-start', '40px')
				await page.evaluate(() => document.fonts.ready)
				// Measure the document, not kanban/gantt's intended internal scrolling.
				expect(await page.evaluate(() => {
					const root = document.documentElement
					return root.scrollWidth - root.clientWidth
				})).toBe(0)
			})
		}
	}
})
