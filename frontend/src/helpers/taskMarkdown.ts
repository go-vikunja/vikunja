import type {ITask} from '@/modelTypes/ITask'
import {DOMParser as ProseMirrorDOMParser} from '@tiptap/pm/model'
import {defaultMarkdownSerializer, MarkdownSerializer, schema} from '@tiptap/pm/markdown'

type MarkdownTask = Pick<ITask, 'title' | 'description'>

const pmDOMParser = ProseMirrorDOMParser.fromSchema(schema)

// Non-strict mode: unknown nodes render their text content instead of throwing
const markdownSerializer = new MarkdownSerializer(
	defaultMarkdownSerializer.nodes,
	defaultMarkdownSerializer.marks,
	{strict: false},
)

// Placeholders for task checkboxes — replaced after serialization
// because the serializer would escape the brackets
const UNCHECKED_MARKER = '\u{FFFC}\u{2610}'
const CHECKED_MARKER = '\u{FFFC}\u{2611}'

/**
 * Pre-processes TipTap HTML so the default ProseMirror schema can parse it.
 */
function preprocessHTML(html: string): string {
	const doc = new DOMParser().parseFromString(html, 'text/html')

	for (const item of doc.querySelectorAll('li[data-type="taskItem"]')) {
		const checked = item.getAttribute('data-checked') === 'true'
		const prefix = checked ? CHECKED_MARKER : UNCHECKED_MARKER
		const firstBlock = item.querySelector('p, div')
		if (firstBlock) {
			firstBlock.insertBefore(doc.createTextNode(prefix), firstBlock.firstChild)
		} else {
			item.insertBefore(doc.createTextNode(prefix), item.firstChild)
		}
		item.removeAttribute('data-type')
		item.removeAttribute('data-checked')
	}

	for (const list of doc.querySelectorAll('ul[data-type="taskList"]')) {
		list.removeAttribute('data-type')
	}

	for (const table of doc.querySelectorAll('table')) {
		const rows: string[] = []
		for (const tr of table.querySelectorAll('tr')) {
			const cells: string[] = []
			for (const cell of tr.querySelectorAll('td, th')) {
				cells.push((cell.textContent || '').trim())
			}
			rows.push(cells.join(' | '))
		}
		const replacement = doc.createElement('div')
		for (const row of rows) {
			const p = doc.createElement('p')
			p.textContent = row
			replacement.appendChild(p)
		}
		table.replaceWith(replacement)
	}

	return doc.body.innerHTML
}

function htmlToMarkdown(html: string): string {
	if (!html || html === '<p></p>') {
		return ''
	}

	const preprocessed = preprocessHTML(html)
	const dom = new DOMParser().parseFromString(preprocessed, 'text/html')
	const doc = pmDOMParser.parse(dom.body)

	return markdownSerializer.serialize(doc)
		.replace(new RegExp(UNCHECKED_MARKER, 'g'), '[ ] ')
		.replace(new RegExp(CHECKED_MARKER, 'g'), '[x] ')
		.trim()
}

/**
 * Format a task as Markdown (title as heading + description).
 */
export function formatTaskAsMarkdown(task: MarkdownTask): string {
	const parts: string[] = [`# ${task.title.trim()}`]

	const description = htmlToMarkdown(task.description)
	if (description) {
		parts.push('', description)
	}

	return parts.join('\n')
}
