import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {UserProjectFactory} from '../../factories/users_project'
import {TeamProjectFactory} from '../../factories/team_project'
import {LinkShareFactory} from '../../factories/link_sharing'
import {login} from '../../support/authenticateUser'
import {createProjects} from '../project/prepareProjects'

for (const kind of ['user', 'team'] as const) {
	test(`adds, changes and removes a project ${kind} share`, async ({authenticatedPage: page}) => {
		const users = await UserFactory.create(2)
		await createProjects(1)
		const [team] = await TeamFactory.create(1, {id: 1, name: 'Sharing Team'})
		await TeamMemberFactory.create(1, {team_id: team.id, user_id: 1, admin: true})
		const name = kind === 'user' ? users[1].username : team.name
		await page.goto('/projects/1/settings/share')
		const section = page.getByRole('heading', {name: `Shared with these ${kind}s`, exact: true}).locator('..')
		await section.getByRole('combobox', {name: `Search for a ${kind} to share this project with`}).pressSequentially(name, {delay: 10})
		await expect(section.locator('.search-results').getByRole('option').first()).toContainText(name)
		await section.locator('.search-results').getByRole('option').first().click()
		const share = section.getByRole('button', {name: 'Share', exact: true})
		await share.focus()
		await share.press('Enter')
		const row = section.getByRole('row').filter({hasText: name})
		await expect(row).toBeVisible()
		await expect(share).toBeFocused()
		if (kind === 'team') {
			await expect(row.getByRole('link', {name})).toHaveAttribute('href', `/teams/${team.id}/edit`)
		}
		await row.getByRole('combobox').selectOption('1')
		await expect(row.locator('td.type')).toHaveText('Read & write')
		await page.reload()
		await expect(row.locator('td.type')).toHaveText('Read & write')
		await row.getByRole('button', {name: `Remove this ${kind}`}).click()
		const removeDialog = page.getByRole('dialog', {name: `Remove a ${kind} from the project`})
		await removeDialog.getByRole('button', {name: 'Cancel', exact: true}).click()
		await expect(removeDialog).toHaveCount(0)
		await expect(row).toBeVisible()
		await row.getByRole('button', {name: `Remove this ${kind}`}).click()
		await page.locator('dialog[open]').getByRole('button', {name: 'Do it!'}).click()
		await expect(row).toHaveCount(0)
		await page.reload()
		await expect(row).toHaveCount(0)
	})
}

for (const kind of ['user', 'team'] as const) {
	test(`clears ${kind} share search results when the search is cleared`, async ({authenticatedPage: page}) => {
		const users = await UserFactory.create(2)
		await createProjects(1)
		const [team] = await TeamFactory.create(1, {id: 1, name: 'Sharing Team'})
		await TeamMemberFactory.create(1, {team_id: team.id, user_id: 1, admin: true})
		const name = kind === 'user' ? users[1].username : team.name
		await page.goto('/projects/1/settings/share')
		const section = page.getByRole('heading', {name: `Shared with these ${kind}s`, exact: true}).locator('..')
		const search = section.getByRole('combobox', {name: `Search for a ${kind} to share this project with`})
		await search.pressSequentially(name, {delay: 10})
		await expect(section.locator('.search-results').getByRole('option')).not.toHaveCount(0)
		// fill('') skips the keyup handler the search binds to; select-all + Backspace is a real keystroke.
		await search.press('ControlOrMeta+a')
		await search.press('Backspace')
		await expect(section.locator('.search-results').getByRole('option')).toHaveCount(0)
	})
}

test('read-only members cannot manage shares', async ({page, apiContext}) => {
	const [, reader] = await UserFactory.create(2)
	await createProjects(1)
	await UserProjectFactory.create(1, {project_id: 1, user_id: reader.id, permission: 0})
	await login(page, apiContext, reader)
	await page.goto('/projects/1/settings/share')
	const section = page.getByRole('heading', {name: 'Shared with these users', exact: true}).locator('..')
	await expect(section.getByRole('row').filter({hasText: reader.username}).getByText('You', {exact: true})).toBeVisible()
	await expect(section.getByRole('combobox', {name: 'Search for a user to share this project with'})).toHaveCount(0)
	await expect(section.getByRole('button', {name: 'Share', exact: true})).toHaveCount(0)
	await expect(section.getByRole('button', {name: 'Remove this user'})).toHaveCount(0)
})

test.describe('Link shares of a user losing access', () => {
	async function createShares(memberId: number) {
		await LinkShareFactory.create(1, {id: 1, name: 'Client link', permission: 1, shared_by_id: memberId})
		await LinkShareFactory.create(1, {id: 2, name: 'Owner link', shared_by_id: 1}, false)
	}

	async function storedShareNames(apiContext, token: string) {
		const response = await apiContext.get('/api/v2/projects/1/shares', {
			headers: {Authorization: `Bearer ${token}`},
		})
		expect(response.ok()).toBe(true)
		return (await response.json()).items.map(share => share.name).sort()
	}

	for (const revoke of [true, false]) {
		test(`removes a user ${revoke ? 'and revokes' : 'but keeps'} their link shares`, async ({
			authenticatedPage: page,
			apiContext,
			userToken,
		}) => {
			const [, member] = await UserFactory.create(2)
			await createProjects(1)
			await UserProjectFactory.create(1, {project_id: 1, user_id: member.id, permission: 2})
			await createShares(member.id)
			await page.goto('/projects/1/settings/share')
			const row = page
				.getByRole('heading', {name: 'Shared with these users', exact: true})
				.locator('..')
				.getByRole('row')
				.filter({hasText: member.username})
			await row.getByRole('button', {name: 'Remove this user'}).click()
			const dialog = page.getByRole('dialog', {name: 'Remove a user from the project'})
			const listed = dialog.getByTestId('revocable-link-share')
			await expect(listed).toHaveCount(1)
			await expect(listed).toContainText('Client link')
			await expect(listed).toContainText('Read & write')
			const checkbox = dialog.getByRole('checkbox', {name: 'Also revoke this link share'})
			await expect(checkbox).not.toBeChecked()
			if (revoke) {
				await dialog.getByText('Also revoke this link share').click()
				await expect(checkbox).toBeChecked()
			}
			await dialog.getByRole('button', {name: 'Do it!'}).click()
			await expect(row).toHaveCount(0)
			await expect
				.poll(() => storedShareNames(apiContext, userToken))
				.toEqual(revoke ? ['Owner link'] : ['Client link', 'Owner link'])
		})
	}

	test('shows no revoke option when the user created no link shares', async ({authenticatedPage: page}) => {
		const [, member] = await UserFactory.create(2)
		await createProjects(1)
		await UserProjectFactory.create(1, {project_id: 1, user_id: member.id, permission: 2})
		await LinkShareFactory.create(1, {name: 'Owner link', shared_by_id: 1})
		await page.goto('/projects/1/settings/share')
		await page
			.getByRole('row')
			.filter({hasText: member.username})
			.getByRole('button', {name: 'Remove this user'})
			.click()
		const dialog = page.getByRole('dialog', {name: 'Remove a user from the project'})
		await expect(dialog.getByText('their task assignments')).toBeVisible()
		await expect(dialog.getByTestId('revocable-link-share')).toHaveCount(0)
		await expect(dialog.getByRole('checkbox')).toHaveCount(0)
	})

	test('removes a team member and revokes their link shares', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		const [, member] = await UserFactory.create(2)
		await createProjects(1)
		await TeamFactory.create(1, {id: 1, name: 'Sharing Team'})
		await TeamMemberFactory.create(1, {team_id: 1, user_id: 1, admin: true})
		await TeamMemberFactory.create(1, {team_id: 1, user_id: member.id}, false)
		await TeamProjectFactory.create(1, {team_id: 1, project_id: 1, permission: 2})
		await createShares(member.id)
		await page.goto('/teams/1/edit')
		await page
			.getByRole('row')
			.filter({hasText: member.username})
			.getByRole('button', {name: 'Remove a user from the team'})
			.click()
		const dialog = page.getByRole('dialog', {name: 'Remove a user from the team'})
		await expect(dialog.getByTestId('revocable-link-share')).toHaveText([/Client link/])
		await dialog.getByText('Also revoke this link share').click()
		await dialog.getByRole('button', {name: 'Do it!'}).click()
		await expect(page.getByRole('row').filter({hasText: member.username})).toHaveCount(0)
		await expect.poll(() => storedShareNames(apiContext, userToken)).toEqual(['Owner link'])
	})
})
