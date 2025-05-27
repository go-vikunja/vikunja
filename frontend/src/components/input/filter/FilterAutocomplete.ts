import {Extension} from '@tiptap/core'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {VueRenderer} from '@tiptap/vue-3'
import tippy from 'tippy.js'

import FilterCommandsList from './FilterCommandsList.vue'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	FILTER_OPERATORS_REGEX,
	LABEL_FIELDS,
	PROJECT_FIELDS,
} from '@/helpers/filters'

import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import UserService from '@/services/user'
import ProjectUserService from '@/services/projectUsers'

export interface FilterAutocompleteOptions {
	projectId?: number
}

export default Extension.create<FilterAutocompleteOptions>({
	name: 'filterAutocomplete',

	addOptions() {
		return {
			projectId: undefined,
		}
	},

	addProseMirrorPlugins() {
		const labelStore = useLabelStore()
		const projectStore = useProjectStore()
		const userService = new UserService()
		const projectUserService = new ProjectUserService()

		let popup: any
		let component: VueRenderer
		let currentAutocompleteContext: any = null

		const updateAutocomplete = async (view: any, force = false) => {
			const {from} = view.state.selection
			const text = view.state.doc.textContent
			const textUpToCursor = text.substring(0, from)

			// Check if we're in an autocomplete context
			let autocompleteContext = null
			let fieldType: 'labels' | 'assignees' | 'projects' | null = null

			for (const field of AUTOCOMPLETE_FIELDS) {
				const pattern = new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()]*)?$', 'ig')
				const match = pattern.exec(textUpToCursor)

				if (match) {
					const [, prefix, , , keyword = ''] = match
					
					let search = keyword.trim()
					const operator = match[0].match(new RegExp(FILTER_OPERATORS_REGEX))?.[0]
					if (operator === 'in' || operator === '?=') {
						const keywords = keyword.split(',')
						search = keywords[keywords.length - 1].trim()
					}
					
					autocompleteContext = {
						field,
						prefix,
						keyword,
						search,
						startPos: match.index + prefix.length,
						endPos: match.index + prefix.length + keyword.length,
					}

					if (LABEL_FIELDS.includes(field)) {
						fieldType = 'labels'
					} else if (ASSIGNEE_FIELDS.includes(field)) {
						fieldType = 'assignees'
					} else if (PROJECT_FIELDS.includes(field)) {
						fieldType = 'projects'
					}
					break
				}
			}

			// If no autocomplete context or same context, and not forced, return
			if (!force && currentAutocompleteContext === autocompleteContext) {
				return
			}

			currentAutocompleteContext = autocompleteContext

			// Hide popup if no context
			if (!autocompleteContext || !fieldType) {
				if (popup) {
					popup.hide()
				}
				return
			}

			// Get suggestions based on field type
			let suggestions: any[] = []

			try {
				if (fieldType === 'labels') {
					suggestions = labelStore.filterLabelsByQuery([], autocompleteContext.search)
				} else if (fieldType === 'assignees') {
					if (this.options.projectId) {
						suggestions = await projectUserService.getAll({projectId: this.options.projectId} as any, {s: autocompleteContext.search})
					} else {
						suggestions = await userService.getAll({} as any, {s: autocompleteContext.search})
					}
					// For assignees, show suggestions even with empty search, but limit if we have many
					if (autocompleteContext.search === '' && suggestions.length > 10) {
						suggestions = suggestions.slice(0, 10)
					}
				} else if (fieldType === 'projects' && !this.options.projectId) {
					suggestions = projectStore.searchProject(autocompleteContext.search)
				}
			} catch (error) {
				console.error('Error fetching suggestions:', error)
				suggestions = []
			}

			// Transform suggestions to match CommandsList format
			const items = suggestions.map(item => ({
				id: item.id,
				title: fieldType === 'assignees' ? item.username : item.title,
				description: fieldType === 'assignees' ? `${item.name || item.username}` : item.title,
				item,
				fieldType,
				context: autocompleteContext,
			}))

			if (items.length === 0) {
				if (popup) {
					popup.hide()
				}
				return
			}

			// Create or update component
			if (!component) {
				component = new VueRenderer(FilterCommandsList, {
					props: {
						items,
						command: (item: any) => {
							// Handle selection
							const newValue = item.fieldType === 'assignees' ? item.item.username : item.item.title
							const currentText = view.state.doc.textContent
							
							// Find the search term and replace it
							const searchStart = currentText.lastIndexOf(item.context.search, from)
							if (searchStart !== -1) {
								const transaction = view.state.tr.replaceWith(
									searchStart,
									searchStart + item.context.search.length,
									view.state.schema.text(newValue)
								)
								view.dispatch(transaction)
							}

							if (popup) {
								popup.hide()
							}
						},
					},
					editor: this.editor,
				})
			} else {
				component.updateProps({
					items,
				})
			}

			// Create or update popup
			if (!popup) {
				popup = tippy(view.dom, {
					getReferenceClientRect: () => {
						const coords = view.coordsAtPos(from)
						return {
							width: 0,
							height: 0,
							top: coords.top,
							bottom: coords.bottom,
							left: coords.left,
							right: coords.left,
						}
					},
					appendTo: () => document.body,
					content: component.element,
					showOnCreate: true,
					interactive: true,
					trigger: 'manual',
					placement: 'bottom-start',
				})[0]
			} else {
				popup.show()
			}
		}

		return [
			new Plugin({
				key: new PluginKey('filterAutocomplete'),
				view() {
					return {
						update: (view, prevState) => {
							// Only update if the document or selection changed
							if (
								!prevState ||
								!view.state.doc.eq(prevState.doc) ||
								!view.state.selection.eq(prevState.selection)
							) {
								setTimeout(() => updateAutocomplete(view), 0)
							}
						},
						destroy() {
							if (popup) {
								popup.destroy()
							}
							if (component) {
								component.destroy()
							}
						},
					}
				},
				props: {
					handleKeyDown(view, event) {
						if (!popup || !popup.state.isVisible) {
							return false
						}

						// Forward key events to the component
						if (component?.ref?.onKeyDown) {
							return component.ref.onKeyDown({event})
						}

						return false
					},
				},
			}),
		]
	},
})
