import { TaskItem } from '@tiptap/extension-list'
import { nanoid } from 'nanoid'

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
					// Always ensure we have an ID
					const id = attributes.taskId || nanoid(8)
					return {
						'data-task-id': id,
					}
				},
			},
		}
	},
})
