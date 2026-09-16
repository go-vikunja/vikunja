import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'

test('creates a scoped MCP token and shows connection instructions once', async ({authenticatedPage: page}) => {
	await gotoUserSettings(page, 'mcp')
	await expect(page.locator('.card-header-title')).toBeVisible()
	await expect(page.getByRole('link', {name: 'More information about MCP in Vikunja'})).toHaveAttribute('href', 'https://vikunja.io/help/mcp/')
	await expect(page.getByLabel('Which client do you use?')).toHaveCount(0)
	const endpoint = await page.getByLabel('MCP endpoint').inputValue()
	await page.getByRole('button', {name: 'Create a token'}).click()
	await expect(page.getByRole('checkbox', {name: /access$/})).toBeChecked()
	await expect(page.getByRole('checkbox', {name: /access$/})).toBeDisabled()
	const created = page.waitForResponse(r => r.url().endsWith('/tokens') && r.request().method() === 'PUT')
	await page.getByRole('button', {name: 'Create token', exact: true}).click()
	const response = await created
	expect(response.ok()).toBeTruthy()
	const token = await response.json()
	expect(token.permissions.mcp).toEqual(['access'])
	expect(token.permissions.tasks).toContain('update')
	expect(token.permissions.other).toContain('users')
	await page.getByLabel('Which client do you use?').selectOption('claudeCode')
	await expect(page.locator('pre')).toContainText(token.token)
	await expect(page.locator('pre')).toContainText(endpoint)
	await expect(page.locator('pre')).toContainText('claude mcp add --transport http')
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
	await expect(page.getByLabel('Which client do you use?')).toHaveCount(0)
	await expect(page.locator('tbody')).toContainText('MCP')
	await page.reload()
	await expect(page.locator('body')).not.toContainText(token.token)
	await page.getByRole('button', {name: 'Delete', exact: true}).click()
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(page.locator('tbody tr')).toHaveCount(0)
})
