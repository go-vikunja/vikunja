import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {UserProjectFactory} from '../../factories/users_project'
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
		const removeDialog = page.getByRole('dialog', {name: `Remove a ${kind} from the List`})
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
