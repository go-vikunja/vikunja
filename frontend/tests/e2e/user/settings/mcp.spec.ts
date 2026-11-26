import {createHash, randomBytes} from 'node:crypto'
import type {APIRequestContext, Page} from '@playwright/test'
import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {apiV2Url} from '../../../support/apiUrl'
import {ApiTokenFactory} from '../../../factories/api_token'

function mcpCard(page: Page) {
	return page.locator('.card').filter({has: page.getByLabel('MCP endpoint')})
}

function connectedAppsCard(page: Page) {
	return page.locator('.card').filter({has: page.getByText('Connected apps', {exact: true})})
}

async function v2(apiContext: APIRequestContext, token: string, method: 'get' | 'post', path: string, options: {
	data?: unknown,
	form?: Record<string, string>,
} = {}) {
	const response = await apiContext[method](`${apiV2Url()}/${path}`, {
		headers: {Authorization: `Bearer ${token}`},
		...options,
	})
	expect(response.ok(), await response.text()).toBeTruthy()
	return response.json()
}

// Runs the real authorize and token exchange, so the grant carries the permissions the server derives from the scope.
async function grantReadAccess(apiContext: APIRequestContext, userToken: string, clientId: string) {
	const {endpoint} = await v2(apiContext, userToken, 'get', 'mcp/info')
	const verifier = randomBytes(32).toString('base64url')
	const redirectUri = 'http://127.0.0.1:33418/callback'
	const {code} = await v2(apiContext, userToken, 'post', 'oauth/authorize', {data: {
		response_type: 'code',
		client_id: clientId,
		redirect_uri: redirectUri,
		code_challenge: createHash('sha256').update(verifier).digest('base64url'),
		code_challenge_method: 'S256',
		scope: 'mcp:read',
		resource: endpoint,
	}})
	await v2(apiContext, userToken, 'post', 'oauth/token', {form: {
		grant_type: 'authorization_code',
		code,
		client_id: clientId,
		redirect_uri: redirectUri,
		code_verifier: verifier,
		resource: endpoint,
	}})
}

test('creates a scoped MCP token and shows connection instructions once', async ({authenticatedPage: page}) => {
	await gotoUserSettings(page, 'mcp')
	const card = mcpCard(page)
	await expect(card.locator('.card-header-title')).toBeVisible()
	await expect(page.getByRole('link', {name: 'More information about MCP in Vikunja'})).toHaveAttribute('href', 'https://vikunja.io/help/mcp/')
	const endpoint = await page.getByLabel('MCP endpoint').inputValue()
	await expect(card.getByRole('heading', {name: 'Connect with OAuth'})).toBeVisible()
	await page.getByLabel('Which client do you use?').selectOption('claudeCode')
	await expect(card.locator('pre')).toHaveText(`claude mcp add --transport http vikunja '${endpoint}'`)
	await page.getByRole('button', {name: 'Create a token'}).click()
	await expect(page.getByRole('checkbox', {name: /access$/})).toBeChecked()
	await expect(page.getByRole('checkbox', {name: /access$/})).toBeDisabled()
	const created = page.waitForResponse(r => r.url().endsWith('/tokens') && r.request().method() === 'POST')
	await page.getByRole('button', {name: 'Create token', exact: true}).click()
	const response = await created
	expect(response.ok()).toBeTruthy()
	const token = await response.json()
	expect(token.permissions.mcp).toEqual(['access'])
	expect(token.permissions.tasks).toContain('update')
	expect(token.permissions.other).toContain('users')
	await expect(card.getByRole('heading', {name: 'Connect with OAuth'})).toHaveCount(0)
	await expect(page.getByLabel('Which client do you use?')).toHaveCount(1)
	await page.getByLabel('Which client do you use?').selectOption('claudeCode')
	await expect(card.locator('pre')).toContainText(token.token)
	await expect(card.locator('pre')).toContainText(endpoint)
	await expect(card.locator('pre')).toContainText('claude mcp add --transport http')
	for (const [client, name, url] of [
		['claudeCode', 'Claude Code', 'https://code.claude.com/docs/en/mcp'],
		['codex', 'Codex', 'https://learn.chatgpt.com/docs/extend/mcp?surface=cli'],
		['claudeDesktop', 'Claude Desktop / claude.ai', 'https://claude.com/docs/connectors/custom/remote-mcp'],
		['mistral', 'Mistral Vibe', 'https://docs.mistral.ai/vibe/work/connectors/mcp-connectors'],
		['chatgpt', 'ChatGPT', 'https://developers.openai.com/api/docs/guides/developer-mode'],
	]) {
		await page.getByLabel('Which client do you use?').selectOption(client)
		await expect(page.getByRole('link', {name: `${name} setup guide`, exact: true})).toHaveAttribute('href', url)
	}
	await page.getByLabel('Which client do you use?').selectOption('mistral')
	await expect(page.locator('.mcp-steps')).toContainText('Context → Connectors → Add Connector → Add Custom Connector')
	await page.getByRole('button', {name: 'Done', exact: true}).click()
	await expect(card.getByRole('heading', {name: 'Connect with OAuth'})).toBeVisible()
	await expect(card.locator('tbody')).toContainText('MCP')
	await page.reload()
	await expect(page.locator('body')).not.toContainText(token.token)
	await card.getByRole('button', {name: 'Delete', exact: true}).click()
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(card.locator('tbody tr')).toHaveCount(0)
	await page.reload()
	await expect(card.locator('tbody tr')).toHaveCount(0)
})

test('lists oauth grants as connected apps only and revokes one', async ({authenticatedPage: page, apiContext, currentUser, userToken}) => {
	await ApiTokenFactory.create(1, {
		title: 'Example Assistant',
		owner_id: currentUser.id,
		oauth_client_id: 'https://assistant.example.com/oauth/client.json',
		oauth_resource: 'https://vikunja.example/api/v2/mcp',
		permissions: JSON.stringify({
			mcp: ['access'],
			tasks: ['read_all', 'create'],
		}),
	})
	await grantReadAccess(apiContext, userToken, 'e2e-desktop-client')
	await v2(apiContext, userToken, 'post', 'tokens', {data: {
		title: 'Made by hand',
		permissions: {mcp: ['access']},
		expires_at: '2099-01-01T00:00:00Z',
	}})

	await gotoUserSettings(page, 'mcp')
	const apps = connectedAppsCard(page)
	await expect(apps.locator('tbody tr')).toHaveCount(2)
	const assistant = apps.locator('tbody tr').filter({hasText: 'Example Assistant'})
	await expect(assistant).toContainText('Published by assistant.example.com')
	await expect(assistant).toContainText('Read & write')
	const desktop = apps.locator('tbody tr').filter({hasText: 'e2e-desktop-client'})
	await expect(desktop).toContainText('Read only')
	await expect(desktop).not.toContainText('Published by')
	await expect(mcpCard(page).locator('tbody tr')).toHaveCount(1)
	await expect(mcpCard(page).locator('tbody')).toContainText('Made by hand')

	await gotoUserSettings(page, 'api-tokens')
	await expect(page.locator('tbody tr')).toHaveCount(1)
	await expect(page.locator('tbody')).toContainText('Made by hand')

	await gotoUserSettings(page, 'mcp')
	await desktop.getByRole('button', {name: 'Revoke'}).click()
	await expect(page.locator('dialog')).toContainText('e2e-desktop-client will no longer be able to access your account.')
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(apps.locator('tbody tr')).toHaveCount(1)
	await page.reload()
	await expect(apps.locator('tbody tr')).toHaveCount(1)
	await expect(apps.locator('tbody')).toContainText('Example Assistant')
	const {items} = await v2(apiContext, userToken, 'get', 'tokens?kind=oauth')
	expect(items.map((item: {title: string}) => item.title)).toEqual(['Example Assistant'])
})

test('shows an empty list of connected apps', async ({authenticatedPage: page}) => {
	await gotoUserSettings(page, 'mcp')
	await expect(connectedAppsCard(page)).toContainText('You have not connected any apps yet.')
})
