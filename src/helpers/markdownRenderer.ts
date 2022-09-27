import {marked} from 'marked'
import hljs from 'highlight.js/lib/common'

export function setupMarkdownRenderer(checkboxId: string) {
	const renderer = new marked.Renderer()
	const linkRenderer = renderer.link

	let checkboxNum = -1
	marked.use({
		renderer: {
			image(src: string, title: string, text: string) {

				title = title ? ` title="${title}` : ''

				// If the url starts with the api url, the image is likely an attachment and
				// we'll need to download and parse it properly.
				if (src.slice(0, window.API_URL.length + 7) === `${window.API_URL}/tasks/`) {
					return `<img data-src="${src}" alt="${text}" ${title} class="attachment-image"/>`
				}

				return `<img src="${src}" alt="${text}" ${title}/>`
			},
			checkbox(checked: boolean) {
				let checkedString = ''
				if (checked) {
					checkedString = 'checked'
				}

				checkboxNum++
				return `<input type="checkbox" data-checkbox-num="${checkboxNum}" ${checkedString} class="text-checkbox-${checkboxId}"/>`
			},
			link(href: string, title: string, text: string) {
				const isLocal = href.startsWith(`${location.protocol}//${location.hostname}`)
				const html = linkRenderer.call(renderer, href, title, text)
				return isLocal ? html : html.replace(/^<a /, '<a target="_blank" rel="noreferrer noopener nofollow" ')
			},
		},
		highlight(code: string, language: string) {
			const validLanguage = hljs.getLanguage(language) ? language : 'plaintext'
			return hljs.highlight(code, {language: validLanguage}).value
		},
	})
	
	return renderer
}