import {test, expect} from '../../support/fixtures'
import type {Page, Locator} from '@playwright/test'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TimeEntryFactory} from '../../factories/time_entry'
import {LicenseFactory} from '../../factories/license'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'

// Pick a project in the form's project picker. Waits for the project store to
// hydrate (the sidebar shows it) before searching so the result is there.
async function selectProject(page: Page, form: Locator, title: string) {
	await expect(page.locator('.menu-container').getByText(title)).toBeVisible()
	const input = form.locator('.multiselect').first().locator('input')
	await input.click()
	// pressSequentially (not fill) so the multiselect's @keyup search fires.
	await input.pressSequentially(title, {delay: 30})
	await form.locator('.search-result-button').filter({hasText: title}).first().click()
}

// Pick a task in the form's task picker (the second multiselect, after project).
async function selectTask(form: Locator, title: string) {
	const input = form.locator('.multiselect').nth(1).locator('input')
	await input.click()
	await input.pressSequentially(title, {delay: 30})
	await form.locator('.search-result-button').filter({hasText: title}).first().click()
}

// Open the time-tracking section on a task detail page.
async function openTaskTimeTracking(page: Page, taskId: number): Promise<Locator> {
	await page.goto(`/tasks/${taskId}`)
	await page.locator('[data-cy="taskTrackTimeAction"]').click()
	const section = page.locator('.task-time-tracking')
	await expect(section).toBeVisible()
	return section
}

test.describe('Time tracking', () => {
	test.describe('with the feature licensed', () => {
		test.beforeEach(async () => {
			await LicenseFactory.enable(['time_tracking'])
		})

		test.afterEach(async () => {
			await LicenseFactory.disable()
		})

		test('shows the page and the sidebar entry', async ({authenticatedPage: page}) => {
			await page.goto('/')
			await expect(page.locator('.menu-container').getByRole('link', {name: 'Time tracking'})).toBeVisible()

			await page.goto('/time-tracking')
			await expect(page.locator('[data-cy="addTimeEntry"]')).toBeVisible()
		})

		test('logs a manual time entry', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {title: 'E2E tracked project'}, false)

			await page.goto('/time-tracking')
			await page.locator('[data-cy="addTimeEntry"]').click()

			const form = page.locator('[data-cy="timeEntryForm"]')
			await expect(form).toBeVisible()

			await selectProject(page, form, 'E2E tracked project')
			// Smart-fill populates both from and to, so the entry is complete.
			await form.locator('[data-cy="smartFill"]').click()
			await form.locator('[data-cy="saveTimeEntry"]').click()

			await expect(page.locator('[data-cy="timeEntry"]').filter({hasText: 'E2E tracked project'})).toBeVisible()
		})

		test('saving with an empty To logs a completed entry, not a running timer', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {title: 'E2E save project'}, false)

			await page.goto('/time-tracking')
			await page.locator('[data-cy="addTimeEntry"]').click()
			const form = page.locator('[data-cy="timeEntryForm"]')
			await selectProject(page, form, 'E2E save project')
			// No smart-fill: leave "To" empty, then Save.
			await form.locator('[data-cy="saveTimeEntry"]').click()

			// The entry is completed (no open-ended "…") and no timer started.
			const entries = page.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).not.toContainText('…')
			await expect(page.locator('[data-cy="timerBadge"]')).not.toBeVisible()
		})

		test('switching from a task to a project logs against the project', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {id: 1, title: 'XOR project'}, false)
			await TaskFactory.create(1, {id: 1, title: 'XOR task', project_id: 1}, false)

			await page.goto('/time-tracking')
			await page.locator('[data-cy="addTimeEntry"]').click()
			const form = page.locator('[data-cy="timeEntryForm"]')

			// Pick a task, then change your mind to a project — the task must be cleared.
			await selectTask(form, 'XOR task')
			await selectProject(page, form, 'XOR project')

			await form.locator('[data-cy="smartFill"]').click()
			await form.locator('[data-cy="saveTimeEntry"]').click()

			const entry = page.locator('[data-cy="timeEntry"]').first()
			await expect(entry).toContainText('XOR project')
			await expect(entry).not.toContainText('XOR task')
		})

		test('starts a timer and stopping it updates the same entry in the list', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {title: 'E2E timer project'}, false)

			await page.goto('/time-tracking')
			await page.locator('[data-cy="addTimeEntry"]').click()

			const form = page.locator('[data-cy="timeEntryForm"]')
			await selectProject(page, form, 'E2E timer project')
			await form.locator('[data-cy="startTimer"]').click()

			const badge = page.locator('[data-cy="timerBadge"]')
			await expect(badge).toBeVisible()

			// The running entry is in the list with an open-ended time range.
			const entries = page.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).toContainText('…')

			await badge.locator('[data-cy="stopTimer"]').click()
			await expect(badge).not.toBeVisible()

			// The same entry is updated in place — end time set, no longer open-ended.
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).not.toContainText('…')
		})

		test('does not show another user\'s readable running timer in the header', async ({
			authenticatedPage: page,
			currentUser,
		}) => {
			const [timerOwner] = await UserFactory.create(1, {id: currentUser.id + 100}, false)
			const [sharedProject] = await ProjectFactory.create(1, {
				id: 1001,
				title: 'Shared active timer project',
				owner_id: timerOwner.id,
			}, false)
			await UserProjectFactory.create(1, {
				project_id: sharedProject.id,
				user_id: currentUser.id,
				permission: 0,
			}, false)
			await TimeEntryFactory.create(1, {
				project_id: sharedProject.id,
				user_id: timerOwner.id,
				end_time: null,
				comment: 'other user running timer',
			}, false)

			const activeTimerHydrated = page.waitForResponse(response =>
				response.request().method() === 'GET' &&
				response.url().includes('/api/v2/time-entries') &&
				response.url().includes('per_page=1'),
			)
			await page.goto('/time-tracking')
			await activeTimerHydrated

			await expect(page.locator('[data-cy="timeEntry"]').filter({hasText: 'other user running timer'})).toBeVisible()
			await expect(page.locator('[data-cy="timerBadge"]')).not.toBeVisible()
		})

		test('hides edit/delete on entries owned by another user', async ({authenticatedPage: page, currentUser}) => {
			const [other] = await UserFactory.create(1, {id: currentUser.id + 100}, false)
			const [shared] = await ProjectFactory.create(1, {id: 2001, title: 'Shared log project', owner_id: other.id}, false)
			await UserProjectFactory.create(1, {project_id: shared.id, user_id: currentUser.id, permission: 0}, false)
			await TimeEntryFactory.create(1, {id: 10, project_id: shared.id, user_id: other.id, comment: 'theirs'}, false)
			await TimeEntryFactory.create(1, {id: 11, project_id: shared.id, user_id: currentUser.id, comment: 'mine'}, false)

			await page.goto('/time-tracking')
			const theirs = page.locator('[data-cy="timeEntry"]').filter({hasText: 'theirs'})
			const mine = page.locator('[data-cy="timeEntry"]').filter({hasText: 'mine'})
			await expect(theirs).toBeVisible()
			await expect(mine).toBeVisible()

			// The current user keeps the controls on their own entry, but not the other's.
			await expect(mine.locator('[data-cy="editTimeEntry"]')).toBeVisible()
			await expect(theirs.locator('[data-cy="editTimeEntry"]')).toHaveCount(0)
			await expect(theirs.locator('[data-cy="deleteTimeEntry"]')).toHaveCount(0)
		})

		test('task detail: logs an entry and toggles the form with the + button', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {title: 'P'}, false)
			await TaskFactory.create(1, {title: 'Tracked task', project_id: 1}, false)

			const section = await openTaskTimeTracking(page, 1)
			const form = section.locator('[data-cy="timeEntryForm"]')

			// No entries yet → the form is shown implicitly.
			await expect(form).toBeVisible()

			await form.locator('[data-cy="smartFill"]').click()
			await form.locator('[data-cy="saveTimeEntry"]').click()
			await expect(section.locator('[data-cy="timeEntry"]')).toHaveCount(1)

			// With an entry, the form collapses behind the + button.
			await expect(form).not.toBeVisible()
			const addButton = section.locator('[data-cy="addTaskTimeEntry"]')
			await expect(addButton).toBeVisible()
			await addButton.click()
			await expect(form).toBeVisible()
		})

		test('task detail: stopping a timer updates the entry in the list', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {title: 'P'}, false)
			await TaskFactory.create(1, {title: 'Timed task', project_id: 1}, false)

			const section = await openTaskTimeTracking(page, 1)
			await section.locator('[data-cy="timeEntryForm"] [data-cy="startTimer"]').click()

			const badge = page.locator('[data-cy="timerBadge"]')
			await expect(badge).toBeVisible()

			const entries = section.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).toContainText('…')

			await badge.locator('[data-cy="stopTimer"]').click()
			await expect(badge).not.toBeVisible()

			await expect(entries).toHaveCount(1)
			await expect(entries.first()).not.toContainText('…')
		})

		test('edits an entry from the list', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {id: 1, title: 'Edit project'}, false)
			await TimeEntryFactory.create(1, {id: 1, project_id: 1, comment: 'original comment'}, false)

			await page.goto('/time-tracking')
			const entries = page.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).toContainText('original comment')

			await entries.first().locator('[data-cy="editTimeEntry"]').click()
			const form = page.locator('[data-cy="timeEntryForm"]')
			const comment = form.locator('[data-cy="timeEntryComment"]')
			await expect(comment).toHaveValue('original comment')
			await comment.fill('edited comment')
			await form.locator('[data-cy="updateTimeEntry"]').click()

			await expect(entries).toHaveCount(1)
			await expect(entries.first()).toContainText('edited comment')
			await expect(entries.first()).not.toContainText('original comment')
		})

		test('deletes an entry from the list', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {id: 1, title: 'Delete project'}, false)
			await TimeEntryFactory.create(1, {id: 1, project_id: 1, comment: 'to be deleted'}, false)

			await page.goto('/time-tracking')
			const entries = page.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(1)

			await entries.first().locator('[data-cy="deleteTimeEntry"]').click()
			await expect(entries).toHaveCount(0)
		})

		test('filters by project, reflected in the url and restored on reload', async ({authenticatedPage: page}) => {
			await ProjectFactory.create(1, {id: 1, title: 'Alpha'}, false)
			await ProjectFactory.create(1, {id: 2, title: 'Beta'}, false)
			await TimeEntryFactory.create(1, {id: 1, project_id: 1, comment: 'alpha entry'}, false)
			await TimeEntryFactory.create(1, {id: 2, project_id: 2, comment: 'beta entry'}, false)

			await page.goto('/time-tracking')
			const entries = page.locator('[data-cy="timeEntry"]')
			await expect(entries).toHaveCount(2)

			// Narrow to project Alpha in the filter modal.
			await page.locator('[data-cy="openTimeTrackingFilters"]').click()
			const dialog = page.locator('dialog[open]')
			const projectInput = dialog.locator('.multiselect').first().locator('input')
			await projectInput.click()
			await projectInput.pressSequentially('Alpha', {delay: 30})
			await dialog.locator('.search-result-button').filter({hasText: 'Alpha'}).first().click()

			// The filter is written to the url.
			await expect(page).toHaveURL(/[?&]project=1\b/)

			// ...and survives a reload (restored from the url): only Alpha's entry.
			await page.reload()
			await expect(entries).toHaveCount(1)
			await expect(entries.first()).toContainText('Alpha')
			await expect(page).toHaveURL(/[?&]project=1\b/)
		})

		test('clearing the date range does not crash the page', async ({authenticatedPage: page}) => {
			await page.goto('/time-tracking')
			// The default range surfaces as "Today" in the toolbar label.
			await expect(page.locator('.time-tracking__range')).toHaveText('Today')

			await page.locator('[data-cy="openTimeTrackingFilters"]').click()
			// Open the range popup (its trigger is the first button in the picker) and clear via Custom.
			await page.locator('dialog[open] .datepicker-with-range-container').getByRole('button').first().click()
			await page.getByRole('button', {name: 'Custom', exact: true}).click()

			// rangeLabel must not call getFullYear on a null date — the page stays alive.
			await expect(page.locator('.time-tracking__range')).toHaveText('Select a range')
			await expect(page.locator('[data-cy="addTimeEntry"]')).toBeVisible()
		})
	})

	test.describe('without the feature licensed', () => {
		test.beforeEach(async () => {
			await LicenseFactory.disable()
		})

		test('hides the sidebar entry and blocks the route', async ({authenticatedPage: page}) => {
			await page.goto('/')
			await expect(page.locator('.menu-container').getByRole('link', {name: 'Time tracking'})).toHaveCount(0)

			await page.goto('/time-tracking')
			await expect(page.locator('[data-cy="addTimeEntry"]')).not.toBeVisible()
		})
	})
})
