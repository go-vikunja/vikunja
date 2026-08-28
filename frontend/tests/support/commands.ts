import type {Locator, Page} from '@playwright/test'
import {readFileSync} from 'fs'
import {join, dirname} from 'path'
import {fileURLToPath} from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

/**
 * Simulates pasting a file from the clipboard into an element
 * @param locator - The element to paste into
 * @param fileName - The name of the file in the fixtures directory
 * @param fileType - The MIME type of the file (default: 'image/png')
 */
export async function pasteFile(locator: Locator, fileName: string, fileType = 'image/png') {
	const filePath = join(__dirname, '../fixtures', fileName)
	const fileBuffer = readFileSync(filePath)
	const base64 = fileBuffer.toString('base64')

	await locator.evaluate((element, {base64Data, name, type}) => {
		// Convert base64 to blob
		const byteCharacters = atob(base64Data)
		const byteNumbers = new Array(byteCharacters.length)
		for (let i = 0; i < byteCharacters.length; i++) {
			byteNumbers[i] = byteCharacters.charCodeAt(i)
		}
		const byteArray = new Uint8Array(byteNumbers)
		const blob = new Blob([byteArray], {type})

		// Create file and paste event
		const file = new File([blob], name, {type})
		const dataTransfer = new DataTransfer()
		dataTransfer.items.add(file)

		const pasteEvent = new ClipboardEvent('paste', {
			bubbles: true,
			cancelable: true,
			clipboardData: dataTransfer,
		})

		element.dispatchEvent(pasteEvent)
	}, {base64Data: base64, name: fileName, type: fileType})
}

/**
 * Simulates pasting HTML/plain text from the clipboard into an element
 */
export async function pasteHtmlFromClipboard(page: Page, locator: Locator, html: string, text: string) {
	await page.evaluate(async ({html, text}) => {
		const clipboardItem = new ClipboardItem({
			'text/html': new Blob([html], {type: 'text/html'}),
			'text/plain': new Blob([text], {type: 'text/plain'}),
		})

		await navigator.clipboard.write([clipboardItem])
	}, {html, text})

	await locator.focus()
	await page.keyboard.press(process.platform === 'darwin' ? 'Meta+V' : 'Control+V')
}

/**
 * Performs a drag and drop operation
 * Note: Playwright has native dragTo() support, so this is just a wrapper for consistency
 * @param source - The source locator to drag from
 * @param target - The target locator to drop onto
 */
export async function dragAndDrop(source: Locator, target: Locator) {
	await source.dragTo(target)
}
