import {test, expect} from '../../support/fixtures'
import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {UserFactory} from '../../factories/user'

test.describe('Team', () => {
	test('Creates a new team', async ({authenticatedPage: page}) => {
		await TeamFactory.truncate()
		await page.goto('/teams')

		const newTeamName = 'New Team'

		await page.locator('a.button').filter({hasText: 'Create a team'}).click()
		await expect(page).toHaveURL(/\/teams\/new/)
		await expect(page.locator('.card-header-title')).toContainText('Create a team')
		await page.locator('input.input').fill(newTeamName)
		await page.locator('.button').filter({hasText: 'Create'}).click()

		await expect(page).toHaveURL(/\/edit/)
		await expect(page.locator('input#teamtext')).toHaveValue(newTeamName)
	})

	test('Shows all teams', async ({authenticatedPage: page}) => {
		await TeamMemberFactory.create(10, {
			team_id: '{increment}',
		})
		const teams = await TeamFactory.create(10, {
			id: '{increment}',
		})

		await page.goto('/teams')

		await expect(page.locator('.teams.box')).not.toBeEmpty()
		for (const t of teams) {
			await expect(page.locator('.teams.box')).toContainText(t.name)
		}
	})

	test('Allows an admin to edit the team', async ({authenticatedPage: page}) => {
		await TeamMemberFactory.create(1, {
			team_id: 1,
			admin: true,
		})
		await TeamFactory.create(1, {
			id: 1,
		})

		await page.goto('/teams/1/edit')
		await page.locator('.card input.input').first().fill('New Team Name')

		await page.locator('.card .button').filter({hasText: 'Save'}).click()

		await expect(page.locator('table.table td').filter({hasText: 'Admin'})).toBeVisible()
		await expect(page.locator('.global-notification')).toContainText('Success')
	})

	test('Does not allow a normal user to edit the team', async ({authenticatedPage: page}) => {
		await TeamMemberFactory.create(1, {
			team_id: 1,
			admin: false,
		})
		await TeamFactory.create(1, {
			id: 1,
		})

		await page.goto('/teams/1/edit')
		await expect(page.locator('.card input.input')).not.toBeVisible()
		await expect(page.locator('table.table td').filter({hasText: 'Member'})).toBeVisible()
	})

	test('Allows an admin to add members to the team', async ({authenticatedPage: page}) => {
		await TeamMemberFactory.create(1, {
			team_id: 1,
			admin: true,
		})
		await TeamFactory.create(1, {
			id: 1,
		})
		const users = await UserFactory.create(5)

		await page.goto('/teams/1/edit')
		const teamMembersCard = page.locator('.card').filter({hasText: 'Team Members'})
		const multiselect = teamMembersCard.locator('.card-content .multiselect')
		const input = multiselect.locator('.input-wrapper input')

		// Use the full username because the /users endpoint requires exact match
		// Use type/pressSequentially instead of fill to properly trigger Vue's input events
		await input.click()
		await input.pressSequentially(users[1].username, {delay: 10})

		// Wait for search results to appear (there's a 200ms debounce in the multiselect)
		await expect(multiselect.locator('.search-results')).toBeVisible({timeout: 5000})
		await multiselect.locator('.search-results').locator('> *').first().click()
		await teamMembersCard.locator('.card-content .button').filter({hasText: 'Add to team'}).click()

		await expect(page.locator('table.table td').filter({hasText: 'Admin'})).toBeVisible()
		// Find the row containing the new member's username
		const newMemberRow = page.locator('table.table tr').filter({hasText: users[1].username})
		await expect(newMemberRow).toBeVisible()
		await expect(newMemberRow).toContainText('Member')
		await expect(page.locator('.global-notification')).toContainText('Success')
	})
})
