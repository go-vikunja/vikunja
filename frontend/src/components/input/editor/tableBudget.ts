import {elementFromString} from '@tiptap/core'

// Mirrors pkg/richtext/table_budget.go: prosemirror-tables materializes every implied cell.
const MAX_TABLE_CELLS = 10_000

export function parseSpan(cell: Element, name: 'colspan' | 'rowspan'): number {
	return Math.max(parseInt(cell.getAttribute(name) ?? '', 10) || 1, 1)
}

// Over-estimates once rowspans are involved: they extend the height and push later rows' cells right.
function impliedTableCells(table: HTMLTableElement): number {
	let maxRowWidth = 0
	let maxRowspan = 0
	let spannedWidth = 0
	for (const row of table.rows) {
		let width = 0
		for (const cell of row.cells) {
			const colspan = parseSpan(cell, 'colspan')
			const rowspan = parseSpan(cell, 'rowspan')
			width += colspan
			if (rowspan > 1) {
				spannedWidth += colspan
				maxRowspan = Math.max(maxRowspan, rowspan)
			}
		}
		maxRowWidth = Math.max(maxRowWidth, width)
	}
	const height = table.rows.length + 1 + maxRowspan
	const width = maxRowspan > 0 ? 2 * maxRowWidth + spannedWidth : maxRowWidth
	return height * width
}

function replaceWithDiv(element: Element) {
	const div = element.ownerDocument.createElement('div')
	div.append(...element.childNodes)
	element.replaceWith(div)
}

export function limitTableCells(body: HTMLElement) {
	// ProseMirror wraps svg/math tr and td into tables as well.
	for (const element of body.querySelectorAll('tr, td, th')) {
		if (element.namespaceURI !== 'http://www.w3.org/1999/xhtml') {
			replaceWithDiv(element)
		}
	}

	let budget = MAX_TABLE_CELLS
	for (const table of body.querySelectorAll('table')) {
		const cells = impliedTableCells(table)
		if (cells <= budget) {
			budget -= cells
			continue
		}
		for (const element of table.querySelectorAll('thead, tbody, tfoot, tr, td, th')) {
			if (element.closest('table') === table) {
				replaceWithDiv(element)
			}
		}
		replaceWithDiv(table)
	}
}

// Parse the returned DOM directly: serializing it back to a string would let the browser re-parse it into new tables.
export function tableLimitedElementFromString(html: string): HTMLElement {
	const body = elementFromString(html)
	limitTableCells(body)
	return body
}
