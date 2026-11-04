import type {APIRequestContext} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {TeamProjectFactory} from '../../factories/team_project'
import {UserProjectFactory} from '../../factories/users_project'
import {LinkShareFactory} from '../../factories/link_sharing'
import {WebhookFactory} from '../../factories/webhook'
import {createProjects} from '../project/prepareProjects'

const DELETE_WEBHOOKS = 'Delete webhooks they created. Unchecked, they move to the project owner and keep working. Your own webhooks are always deleted.'
const REVOKE_LINK_SHARES = 'Revoke link shares they created'

async function createMemberData(projectId: number, memberId: number) {
	await LinkShareFactory.create(1, {id: 1, name: 'Member link', project_id: projectId, shared_by_id: memberId})
	await LinkShareFactory.create(1, {id: 2, name: 'Owner link', project_id: projectId, shared_by_id: 1}, false)
	await WebhookFactory.create(1, {id: 1, project_id: projectId, target_url: 'https://example.com/member', created_by_id: memberId})
}

async function apiGet(apiContext: APIRequestContext, token: string, path: string) {
	const response = await apiContext.get(`/api/v2${path}`, {headers: {Authorization: `Bearer ${token}`}})
	expect(response.ok()).toBe(true)
	return response.json()
}

async function linkShareNames(apiContext: APIRequestContext, token: string, projectId: number) {
	const {items} = await apiGet(apiContext, token, `/projects/${projectId}/shares`)
	return items.map((share: {name: string}) => share.name).sort()
}

async function webhookCreators(apiContext: APIRequestContext, token: string, projectId: number) {
	const {items} = await apiGet(apiContext, token, `/projects/${projectId}/webhooks`)
	return items.map((webhook: {created_by: {id: number}}) => webhook.created_by.id)
}

test.describe('Removing project access', () => {
	test('defaults delete webhooks and keep link shares', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [, member] = await UserFactory.create(2)
		await createProjects(1)
		await UserProjectFactory.create(1, {project_id: 1, user_id: member.id, permission: 2})
		await createMemberData(1, member.id)
		await page.goto('/projects/1/settings/share')
		const row = page
			.getByRole('heading', {name: 'Shared with these users', exact: true})
			.locator('..')
			.getByRole('row')
			.filter({hasText: member.username})
		await row.getByRole('button', {name: 'Remove this user'}).click()
		const dialog = page.getByRole('dialog', {name: 'Remove a user from the project'})
		await expect(dialog.getByRole('checkbox', {name: DELETE_WEBHOOKS})).toBeChecked()
		await expect(dialog.getByRole('checkbox', {name: REVOKE_LINK_SHARES})).not.toBeChecked()
		await dialog.getByRole('button', {name: 'Do it!'}).click()
		await expect(row).toHaveCount(0)
		await expect.poll(() => webhookCreators(apiContext, userToken, 1)).toEqual([])
		await expect.poll(() => linkShareNames(apiContext, userToken, 1)).toEqual(['Member link', 'Owner link'])
	})

	test('keeps webhooks for the owner and revokes link shares', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [, member] = await UserFactory.create(2)
		await createProjects(1)
		await UserProjectFactory.create(1, {project_id: 1, user_id: member.id, permission: 2})
		await createMemberData(1, member.id)
		await page.goto('/projects/1/settings/share')
		const row = page
			.getByRole('heading', {name: 'Shared with these users', exact: true})
			.locator('..')
			.getByRole('row')
			.filter({hasText: member.username})
		await row.getByRole('button', {name: 'Remove this user'}).click()
		const dialog = page.getByRole('dialog', {name: 'Remove a user from the project'})
		await dialog.getByRole('checkbox', {name: DELETE_WEBHOOKS}).uncheck()
		await dialog.getByRole('checkbox', {name: REVOKE_LINK_SHARES}).check()
		await dialog.getByRole('button', {name: 'Do it!'}).click()
		await expect(row).toHaveCount(0)
		await expect.poll(() => webhookCreators(apiContext, userToken, 1)).toEqual([1])
		await expect.poll(() => linkShareNames(apiContext, userToken, 1)).toEqual(['Owner link'])
	})

	test('removes a team member and revokes their link shares', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [, member] = await UserFactory.create(2)
		await createProjects(1)
		await TeamFactory.create(1, {id: 1, name: 'Sharing Team'})
		await TeamMemberFactory.create(1, {team_id: 1, user_id: 1, admin: true})
		await TeamMemberFactory.create(1, {team_id: 1, user_id: member.id}, false)
		await TeamProjectFactory.create(1, {team_id: 1, project_id: 1, permission: 2})
		await createMemberData(1, member.id)
		await page.goto('/teams/1/edit')
		const row = page.getByRole('row').filter({hasText: member.username})
		await row.getByRole('button', {name: 'Remove a user from the team'}).click()
		const dialog = page.getByRole('dialog', {name: 'Remove a user from the team'})
		await dialog.getByRole('checkbox', {name: REVOKE_LINK_SHARES}).check()
		await dialog.getByRole('button', {name: 'Do it!'}).click()
		await expect(row).toHaveCount(0)
		await expect.poll(() => linkShareNames(apiContext, userToken, 1)).toEqual(['Owner link'])
		await expect.poll(() => webhookCreators(apiContext, userToken, 1)).toEqual([])
	})
})

test.describe('Moving a project', () => {
	// Via the API so the ancestor rows exist.
	async function setupMove(apiContext: APIRequestContext, token: string) {
		const [, member] = await UserFactory.create(2)
		const create = async (body: object) => {
			const response = await apiContext.post('/api/v2/projects', {headers: {Authorization: `Bearer ${token}`}, data: body})
			expect(response.ok()).toBe(true)
			return (await response.json()) as {id: number}
		}
		const parent = await create({title: 'Shared Parent'})
		const child = await create({title: 'Moved Child', parent_project_id: parent.id})
		await UserProjectFactory.create(1, {project_id: parent.id, user_id: member.id, permission: 2})
		return {member, parent, child}
	}

	async function clearParentAndSave(page: import('@playwright/test').Page, childId: number) {
		await page.goto(`/projects/${childId}/settings/edit`)
		await expect(page.locator('.multiselect input')).toHaveValue('Shared Parent')
		await page.locator('.multiselect .removal-button').click()
		await page.locator('footer.card-footer .button').filter({hasText: /^Save$/}).click()
	}

	test('asks about webhooks and link shares and applies the choice', async ({authenticatedPage: page, apiContext, userToken}) => {
		const {member, child} = await setupMove(apiContext, userToken)
		await createMemberData(child.id, member.id)
		await clearParentAndSave(page, child.id)
		const dialog = page.getByRole('dialog', {name: 'Some users will lose access'})
		await expect(dialog).toContainText('They created 1 webhook and 1 link share there.')
		await dialog.getByRole('checkbox', {name: DELETE_WEBHOOKS}).uncheck()
		await dialog.getByRole('checkbox', {name: REVOKE_LINK_SHARES}).check()
		await dialog.getByRole('button', {name: 'Do it!'}).click()
		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect.poll(async () => (await apiGet(apiContext, userToken, `/projects/${child.id}`)).parent_project_id).toBe(0)
		await expect.poll(() => webhookCreators(apiContext, userToken, child.id)).toEqual([1])
		await expect.poll(() => linkShareNames(apiContext, userToken, child.id)).toEqual(['Owner link'])
	})

	test('cancelling keeps the project where it was', async ({authenticatedPage: page, apiContext, userToken}) => {
		const {member, parent, child} = await setupMove(apiContext, userToken)
		await createMemberData(child.id, member.id)
		await clearParentAndSave(page, child.id)
		const dialog = page.getByRole('dialog', {name: 'Some users will lose access'})
		await dialog.getByRole('button', {name: 'Cancel', exact: true}).click()
		await expect(dialog).toHaveCount(0)
		expect((await apiGet(apiContext, userToken, `/projects/${child.id}`)).parent_project_id).toBe(parent.id)
		expect(await webhookCreators(apiContext, userToken, child.id)).toEqual([member.id])
	})

	test('moves without asking when nobody loses webhooks or link shares', async ({authenticatedPage: page, apiContext, userToken}) => {
		const {child} = await setupMove(apiContext, userToken)
		await clearParentAndSave(page, child.id)
		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.getByRole('dialog', {name: 'Some users will lose access'})).toHaveCount(0)
		await expect.poll(async () => (await apiGet(apiContext, userToken, `/projects/${child.id}`)).parent_project_id).toBe(0)
	})
})
