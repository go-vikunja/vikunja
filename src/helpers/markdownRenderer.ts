import {marked} from 'marked'
import hljs from 'highlight.js/lib/common'

export function setupMarkdownRenderer(checkboxId: string) {
	const renderer = new marked.Renderer()
	const linkRenderer = renderer.link

	let checkboxNum = -1
	marked.use({
		renderer: {
			image: (src, title, text) => {

				title = title ? ` title="${title}` : ''

				// If the url starts with the api url, the image is likely an attachment and
				// we'll need to download and parse it properly.
				if (src.substr(0, window.API_URL.length + 7) === `${window.API_URL}/tasks/`) {
					return `<img data-src="${src}" alt="${text}" ${title} class="attachment-image"/>`
				}

				return `<img src="${src}" alt="${text}" ${title}/>`
			},
			checkbox: (checked) => {
				if (checked) {
					checked = ' checked="checked"'
				}

				checkboxNum++
				return `<input type="checkbox" data-checkbox-num="${checkboxNum}" ${checked} class="text-checkbox-${checkboxId}"/>`
			},
			link: (href, title, text) => {
				const isLocal = href.startsWith(`${location.protocol}//${location.hostname}`)
				const html = linkRenderer.call(renderer, href, title, text)
				return isLocal ? html : html.replace(/^<a /, '<a target="_blank" rel="noreferrer noopener nofollow" ')
			},
		},
		highlight: function (code, language) {
			const validLanguage = hljs.getLanguage(language) ? language : 'plaintext'
			return hljs.highlight(code, {language: validLanguage}).value
		},
	})
	
	return renderer
}