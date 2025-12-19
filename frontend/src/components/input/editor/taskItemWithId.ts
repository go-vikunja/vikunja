import { TaskItem } from '@tiptap/extension-list'
import { nanoid } from 'nanoid'
import { Plugin, PluginKey } from '@tiptap/pm/state'

const uniqueIdPluginKey = new PluginKey('taskItemUniqueId')

/**
 * Custom TaskItem extension that adds a unique ID to each task item.
 * This fixes the checkbox persistence bug (GitHub #293, #563) by allowing
 * reliable identification of which checkbox was toggled.
 */
export const TaskItemWithId = TaskItem.extend({
	addAttributes() {
		return {
			...this.parent?.(),
			taskId: {
				default: null,
				parseHTML: (element: HTMLElement) => {
					// Preserve existing ID or generate new one
					return element.getAttribute('data-task-id') || nanoid(8)
				},
				renderHTML: (attributes) => {
					return {
						'data-task-id': attributes.taskId || nanoid(8),
					}
				},
			},
		}
	},

	addProseMirrorPlugins() {
		const parentPlugins = this.parent?.() || []

		return [
			...parentPlugins,
			new Plugin({
				key: uniqueIdPluginKey,
				appendTransaction: (transactions, oldState, newState) => {
					// Only process if document changed
					if (!transactions.some(tr => tr.docChanged)) {
						return null
					}

					const seenIds = new Set<string>()
					const duplicates: { pos: number; node: typeof newState.doc.firstChild }[] = []

					// Find all task items and check for duplicate IDs
					newState.doc.descendants((node, pos) => {
						if (node.type.name === 'taskItem') {
							const taskId = node.attrs.taskId

							if (!taskId || seenIds.has(taskId)) {
								// Missing or duplicate ID - needs regeneration
								duplicates.push({ pos, node })
							} else {
								seenIds.add(taskId)
							}
						}
					})

					// If no duplicates, no transaction needed
					if (duplicates.length === 0) {
						return null
					}

					// Create transaction to fix duplicates
					const tr = newState.tr

					for (const { pos, node } of duplicates) {
						tr.setNodeMarkup(pos, undefined, {
							...node.attrs,
							taskId: nanoid(8),
						})
					}

					return tr
				},
			}),
		]
	},
})
