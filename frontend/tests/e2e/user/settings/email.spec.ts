import {randomBytes} from 'node:crypto'
import type {APIRequestContext} from '@playwright/test'
import {test, expect} from '../../../support/fixtures'
import {TEST_PASSWORD} from '../../../support/constants'

async function confirmationLink(request: APIRequestContext, mailpitUrl: string, address: string) {
	let messageId = ''
	await expect.poll(async () => {
		const search = new URL('/api/v1/search', mailpitUrl).href
		const response = await request.get(search, {params: {query: `to:${address}`}})
		expect(response.status()).toBe(200)
		const result: {messages: {ID: string}[]} = await response.json()
		messageId = result.messages[0]?.ID ?? ''
		return messageId
	}, {
		timeout: 15000,
		message: `Confirmation email for ${address} arrives in Mailpit`,
	}).toBeTruthy()
	const response = await request.get(new URL(`/api/v1/message/${messageId}`, mailpitUrl).href)
	expect(response.status()).toBe(200)
	const mail: {Text: string} = await response.json()
	const link = mail.Text.match(/https?:\/\/[^\s<>]+\?userEmailConfirm=[a-zA-Z0-9]+/)?.[0]
	expect(link).toBeDefined()
	return link!
}

test('stores an email update when mail confirmation is disabled', async ({authenticatedPage: page, apiContext}) => {
	await page.goto('/user/settings/email-update')
	await page.locator('#newEmail').fill('changed@example.test')
	await page.locator('#currentPasswordEmail').fill(TEST_PASSWORD)
	await page.getByRole('button', {name: 'Save', exact: true}).click()
	await expect(page.locator('.global-notification')).toContainText('Success')
	await expect(page.locator('#newEmail')).toHaveValue('')
	await expect(page.locator('#currentPasswordEmail')).toHaveValue('')
	await page.reload()
	const response = await apiContext.post('login', {data: {username: 'changed@example.test', password: TEST_PASSWORD}})
	expect(response.ok()).toBe(true)
	expect((await response.json()).token).toBeTruthy()
})

test('keeps the new address pending until it is cancelled', async ({browser, baseURL, request}) => {
	const publicUrl = new URL('/', baseURL).href
	const apiUrl = (process.env.MAILER_API_URL || 'http://127.0.0.1:3457/api/v1').replace(/\/$/, '')
	const apiV2Url = apiUrl.replace(/\/api\/v1$/, '/api/v2')
	const mailpitUrl = process.env.MAILPIT_URL || 'http://127.0.0.1:8025'
	const username = `pending-email-${randomBytes(8).toString('hex')}`
	const email = `${username}@example.com`
	const newEmail = `new-${username}@example.com`
	const password = '12345678'
	const context = await browser.newContext({baseURL: publicUrl})
	try {
		await context.addInitScript(url => {
			localStorage.setItem('API_URL', url)
			window.API_URL = url
		}, apiUrl)
		const page = await context.newPage()
		await page.goto('/register')
		await page.locator('#username').fill(username)
		await page.locator('#email').fill(email)
		await page.locator('#password').fill(password)
		await page.locator('#register-submit').click()
		await expect(page.locator('div.message.success')).toContainText('check your inbox')
		// Login is only possible once the address the account was registered with is confirmed.
		await page.goto(await confirmationLink(request, mailpitUrl, email))
		await page.locator('#username').fill(username)
		await page.locator('#password').fill(password)
		await page.getByRole('button', {name: 'Login', exact: true}).click()
		await expect(page).toHaveURL(publicUrl)

		const login = await request.post(`${apiUrl}/login`, {
			data: {
				username,
				password,
			},
		})
		expect(login.ok()).toBe(true)
		const headers = {Authorization: `Bearer ${(await login.json()).token}`}
		const storedPendingEmail = async () => {
			const response = await request.get(`${apiV2Url}/user`, {headers})
			expect(response.status()).toBe(200)
			return (await response.json()).pending_email ?? ''
		}

		await page.goto('/user/settings/email-update')
		await page.locator('#newEmail').fill(newEmail)
		await page.locator('#currentPasswordEmail').fill(password)
		await page.getByRole('button', {name: 'Save', exact: true}).click()
		const notice = page.locator('.message.warning')
		await expect(notice).toContainText(newEmail)
		expect(await storedPendingEmail()).toBe(newEmail)

		// Saving already sent a confirmation mail, so resending right away hits the one minute cooldown.
		await notice.getByRole('button', {name: 'Resend confirmation email'}).click()
		const errorToast = page.locator('.global-notification .vue-notification.error')
		await expect(errorToast).toContainText('Please wait a minute before requesting another confirmation email.')
		await expect(notice).toBeVisible()
		expect(await storedPendingEmail()).toBe(newEmail)

		await notice.getByRole('button', {name: 'Cancel change'}).click()
		await expect(notice).toHaveCount(0)
		expect(await storedPendingEmail()).toBe('')

		await page.reload()
		await expect(page.locator('#newEmail')).toHaveValue('')
		await expect(page.locator('.message.warning')).toHaveCount(0)
	} finally {
		await context.close()
	}
})
