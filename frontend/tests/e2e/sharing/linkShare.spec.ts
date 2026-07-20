import {test, expect} from '../../support/fixtures'
import {LabelFactory} from '../../factories/labels'
import {LabelTaskFactory} from '../../factories/label_task'
import {LinkShareFactory} from '../../factories/link_sharing'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {createProjects} from '../project/prepareProjects'
import {login, setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD, TEST_PASSWORD_HASH} from '../../support/constants'

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
	// The anonymous link share tests below don't use the `authenticatedPage`
	// fixture (which wires up the API URL via `login()`), so they'd otherwise
	// hit the default `window.API_URL = '/api/v1'` relative path baked into
	// index.html and never reach the API running on a different port.
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
	})

	test('Can view a link share', async ({page, apiContext}) => {
		const {share, project, tasks} = await prepareLinkShare()

		await page.goto(`/share/${share.hash}/auth`)

		await expect(page.locator('h1.title')).toContainText(project.title)
		await expect(page.locator('input.input[placeholder="Add a task…"]')).not.toBeVisible()
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)

		await expect(page).toHaveURL(`/projects/${project.id}/1#share-auth-token=${share.hash}`)
	})

	test('Should work when directly viewing a project with share hash present', async ({page, apiContext}) => {
		const {share, project, tasks} = await prepareLinkShare()

		await page.goto(`/projects/${project.id}/1#share-auth-token=${share.hash}`)

		await expect(page.locator('h1.title')).toContainText(project.title)
		await expect(page.locator('input.input[placeholder="Add a task…"]')).not.toBeVisible()
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)
	})

	test('Should work when directly viewing a task with share hash present', async ({page, apiContext}) => {
		const {share, project, tasks} = await prepareLinkShare()

		await page.goto(`/tasks/${tasks[0].id}#share-auth-token=${share.hash}`)

		await expect(page.locator('h1.title.input')).toContainText(tasks[0].title)
	})

	// Regression test for #2546: a logged-in user opening a public link share URL
	// used to get stuck on an empty NoAuthWrapper shell because the router
	// guard bounced between /share/:hash/auth and the project view. Two
	// underlying issues produced the same symptom:
	//
	//   1. The 1-minute debounce in `checkAuth()` skipped re-parsing the new
	//      link share JWT when the user was already authenticated.
	//   2. `checkAuth()` skipped `setUser()` when the new JWT's `id` matched
	//      the current `info.value.id`. Because users and link shares share
	//      the same numeric id space, a user whose id happened to match the
	//      link share's id would keep the old USER `info.value.type` and
	//      `authLinkShare` would never flip to true.
	//
	// This test forces the id collision scenario so it covers both issues.
	test('Can view a link share while logged in as a user with a colliding id', async ({page, apiContext}) => {
		// Build the link share setup inline so we can pin the share id and
		// the logged-in user id to the same value, reproducing the real-world
		// bug where `info.value.id === jwtUser.id` is a false positive.
		const collidingId = 42

		const [linkShareOwner] = await UserFactory.create(1, {id: 1})
		const projects = await createProjects()
		const tasks = await TaskFactory.create(10, {
			project_id: projects[0].id,
		})
		const [share] = await LinkShareFactory.create(1, {
			id: collidingId,
			project_id: projects[0].id,
			shared_by_id: linkShareOwner.id,
			permission: 0,
		})
		const project = projects[0]

		// Create the logged-in user with the SAME numeric id as the link share.
		// `truncate=false` so the link share owner (id 1) stays put.
		const [loggedInUser] = await UserFactory.create(1, {id: collidingId}, false)

		await login(page, apiContext, loggedInUser)

		await page.goto(`/share/${share.hash}/auth`)

		// Should successfully land on the shared project view instead of
		// bouncing back to /share/:hash/auth forever.
		await expect(page.locator('h1.title')).toContainText(project.title)
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)
		await expect(page).toHaveURL(`/projects/${project.id}/1#share-auth-token=${share.hash}`)
	})
})

test.describe('Link share: label picker', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
	})

	test('explains that new labels cannot be created when typing an unknown label', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		const [task] = await TaskFactory.create(1, {
			project_id: projects[0].id,
		})
		// A label on the task makes the labels field render without clicking "Add Labels" first.
		const [label] = await LabelFactory.create(1)
		await LabelTaskFactory.create(1, {
			task_id: task.id,
			label_id: label.id,
		})
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			permission: 1,
		})

		await page.goto(`/tasks/${task.id}#share-auth-token=${share.hash}`)

		const labelInput = page.locator('.task-view .details.labels-list .multiselect input')
		await expect(labelInput).toBeVisible()
		await labelInput.fill('label-that-does-not-exist')

		const searchResults = page.locator('.task-view .details.labels-list .multiselect .search-results')
		await expect(searchResults.locator('.search-result-hint')).toContainText('New labels can\'t be created from a shared link')
		await expect(searchResults.locator('.is-create-option')).toHaveCount(0)
	})
})

test.describe('Link share: password protection', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
	})

	test('password-protected share rejects wrong password', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			sharing_type: 2,
			password: TEST_PASSWORD_HASH,
			permission: 0,
		})

		await page.goto(`/share/${share.hash}/auth`)

		// The auth form renders only once the backend returns code 13001, so wait
		// for it before trying to type into the password field.
		const passwordInput = page.locator('input#linkSharePassword')
		await expect(passwordInput).toBeVisible()

		await passwordInput.fill('wrong-password')
		// Wait for the auth POST to complete so we can assert the negative
		// outcome without racing the UI.
		const authRejected = page.waitForResponse(r =>
			r.url().includes(`/shares/${share.hash}/auth`) && r.request().method() === 'POST',
		)
		await page.locator('.button').filter({hasText: 'Login'}).click()
		const resp = await authRejected
		expect(resp.status()).toBeGreaterThanOrEqual(400)

		// The user must not be redirected into the shared project view, and
		// the route stays on the link-share auth URL.
		await expect(page).toHaveURL(new RegExp(`/share/${share.hash}/auth`))
		// No project-title heading renders while we're still on the auth route.
		await expect(page.locator('h1.title')).toHaveCount(0)
	})

	test('password-protected share accepts correct password', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		const tasks = await TaskFactory.create(3, {
			project_id: projects[0].id,
		})
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			sharing_type: 2,
			password: TEST_PASSWORD_HASH,
			permission: 0,
		})

		await page.goto(`/share/${share.hash}/auth`)

		const passwordInput = page.locator('input#linkSharePassword')
		await expect(passwordInput).toBeVisible()

		await passwordInput.fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		await expect(page.locator('h1.title')).toContainText(projects[0].title)
		await expect(page.locator('.tasks')).toContainText(tasks[0].title)
		await expect(page).toHaveURL(`/projects/${projects[0].id}/1#share-auth-token=${share.hash}`)
	})
})

test.describe('Link share: permission tiers', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
	})

	test('READ link share hides add-task', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		await TaskFactory.create(3, {
			project_id: projects[0].id,
		})
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			permission: 0,
		})

		await page.goto(`/share/${share.hash}/auth`)

		// Wait for the project view to actually render so the assertion isn't
		// vacuously true during the loading shell.
		await expect(page.locator('h1.title')).toContainText(projects[0].title)
		await expect(page).toHaveURL(`/projects/${projects[0].id}/1#share-auth-token=${share.hash}`)

		await expect(page.locator('.input[placeholder="Add a task…"]')).toHaveCount(0)
	})

	test('READ_WRITE link share shows add-task', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		await TaskFactory.create(3, {
			project_id: projects[0].id,
		})
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			permission: 1,
		})

		await page.goto(`/share/${share.hash}/auth`)

		await expect(page.locator('h1.title')).toContainText(projects[0].title)
		await expect(page).toHaveURL(`/projects/${projects[0].id}/1#share-auth-token=${share.hash}`)

		await expect(page.locator('.input[placeholder="Add a task…"]')).toBeVisible()
	})
})

test.describe('Link share: quick add magic labels', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
	})

	test('creates the task and shows an error when a label cannot be created', async ({page}) => {
		await UserFactory.create(1)
		const projects = await createProjects()
		await TaskFactory.create(1, {
			project_id: projects[0].id,
		})
		const [share] = await LinkShareFactory.create(1, {
			project_id: projects[0].id,
			permission: 1,
		})

		await page.goto(`/share/${share.hash}/auth`)
		await expect(page.locator('h1.title')).toContainText(projects[0].title)

		const addTaskInput = page.locator('.input[placeholder="Add a task…"]')
		await addTaskInput.fill('New task via share *unknownlabel')
		await addTaskInput.press('Enter')

		// Link shares may not create labels: the label is skipped with an error, the task is still created.
		await expect(page.locator('.global-notification')).toContainText('could not be created')
		await expect(page.locator('.tasks')).toContainText('New task via share')
	})
})
