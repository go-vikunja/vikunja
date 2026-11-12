import {getTextSerializersFromSchema} from '@tiptap/core'
import type {Fragment, Mark, Node, Schema} from '@tiptap/pm/model'

const MARK_DELIMITERS: Record<string, string> = {
	bold: '**',
	italic: '*',
	strike: '~~',
	code: '`',
}

function indent(text: string, prefix: string): string {
	return text.split('\n').map((line, i) => (i === 0 || line === '' ? line : prefix + line)).join('\n')
}

function wrapMark(text: string, mark: Mark): string {
	if (mark.type.name === 'link') {
		return `[${text}](${mark.attrs.href})`
	}
	const delimiter = MARK_DELIMITERS[mark.type.name]
	return delimiter ? delimiter + text + delimiter : text
}

/**
 * Serializes copied content as Markdown for the text/plain clipboard part. ProseMirror's
 * default puts a blank line around every block, so a list item (list > item > paragraph)
 * ends up surrounded by several of them and loses its marker.
 */
export function serializeClipboardText(content: Fragment, schema: Schema): string {
	const textSerializers = getTextSerializersFromSchema(schema)

	function inline(parent: Node): string {
		let result = ''
		parent.forEach((child, offset, index) => {
			if (child.isText) {
				result += child.marks.reduce(wrapMark, child.text ?? '')
			} else if (child.type.name === 'hardBreak') {
				result += '\n'
			} else if (child.type.name === 'image') {
				result += `![${child.attrs.alt ?? ''}](${child.attrs.src})`
			} else {
				const serializer = textSerializers[child.type.name]
				result += serializer
					? serializer({node: child, pos: offset, parent, index, range: {from: 0, to: parent.content.size}})
					: child.textContent
			}
		})
		return result
	}

	function list(node: Node): string {
		const start = node.attrs.start ?? 1
		const items: string[] = []
		node.forEach((item, _offset, i) => {
			let marker = '- '
			if (node.type.name === 'orderedList') {
				marker = `${start + i}. `
			} else if (item.type.name === 'taskItem') {
				marker = item.attrs.checked ? '- [x] ' : '- [ ] '
			}
			items.push(marker + indent(blocks(item.content, '\n'), ' '.repeat(marker.length)))
		})
		return items.join('\n')
	}

	function table(node: Node): string {
		const rows: string[] = []
		node.forEach((row, _offset, i) => {
			const cells: string[] = []
			row.forEach(cell => cells.push(blocks(cell.content, ' ').replaceAll('\n', ' ').replaceAll('|', '\\|')))
			rows.push(`| ${cells.join(' | ')} |`)
			if (i === 0) {
				rows.push(`| ${cells.map(() => '---').join(' | ')} |`)
			}
		})
		return rows.join('\n')
	}

	function block(node: Node): string {
		switch (node.type.name) {
			case 'heading':
				return '#'.repeat(node.attrs.level) + ' ' + inline(node)
			case 'blockquote':
				return blocks(node.content).split('\n').map(line => (line === '' ? '>' : '> ' + line)).join('\n')
			case 'codeBlock':
				return '```' + (node.attrs.language ?? '') + '\n' + node.textContent + '\n```'
			case 'horizontalRule':
				return '---'
			case 'bulletList':
			case 'orderedList':
			case 'taskList':
				return list(node)
			case 'table':
				return table(node)
		}

		if (node.isTextblock) {
			return inline(node)
		}
		if (node.isLeaf) {
			return inline(schema.nodes.paragraph.create(null, node))
		}
		return blocks(node.content)
	}

	function blocks(fragment: Fragment, separator = '\n\n'): string {
		if (fragment.firstChild?.isInline) {
			return inline(schema.nodes.paragraph.create(null, fragment))
		}
		const parts: string[] = []
		fragment.forEach(node => parts.push(block(node)))
		return parts.join(separator)
	}

	return blocks(content)
}
