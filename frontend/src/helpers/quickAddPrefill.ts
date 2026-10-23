import type {LocationQuery} from 'vue-router'

import {isValidHttpUrl} from '@/helpers/isValidHttpUrl'

export interface QuickAddPrefill {
	title: string
	url: string
	description: string
}

export const QUICK_ADD_PREFILL_PARAMS = ['title', 'url', 'description'] as const

function firstString(value: LocationQuery[string]): string {
	const first = Array.isArray(value) ? value[0] : value
	return (first ?? '').trim()
}

export function parseQuickAddPrefill(query: LocationQuery): QuickAddPrefill | null {
	const title = firstString(query.title)
	let url = firstString(query.url)
	let description = firstString(query.description)
	if (title === '' && url === '' && description === '') {
		return null
	}

	// Android share sheets put the page url into the shared text instead.
	if (url === '' && isValidHttpUrl(description)) {
		url = description
		description = ''
	}

	return {title, url, description: description === url ? '' : description}
}

function paragraph(content: string | Node): HTMLParagraphElement {
	const p = document.createElement('p')
	p.append(content)
	return p
}

// Built through the DOM so plain text like "use <div>" stays text instead of becoming markup.
export function quickAddPrefillDescription({url, description}: Pick<QuickAddPrefill, 'url' | 'description'>): string {
	const container = document.createElement('div')
	if (isValidHttpUrl(url)) {
		const link = document.createElement('a')
		link.href = url
		link.textContent = url
		container.append(paragraph(link))
	} else if (url !== '') {
		container.append(paragraph(url))
	}
	description.split(/\n+/).filter(line => line !== '').forEach(line => container.append(paragraph(line)))
	return container.innerHTML
}
