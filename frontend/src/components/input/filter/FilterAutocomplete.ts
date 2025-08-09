import {Extension} from '@tiptap/core'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {VueRenderer} from '@tiptap/vue-3'

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
import {computePosition, flip, shift, offset, autoUpdate} from '@floating-ui/dom'

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

		let popupElement: HTMLElement | null = null
		let component: VueRenderer | null = null
		let currentAutocompleteContext: any = null
		let cleanupFloating: (() => void) | null = null

		const virtualReference = {
			getBoundingClientRect: () => ({
				width: 0,
				height: 0,
				x: 0,
				y: 0,
				top: 0,
				left: 0,
				right: 0,
				bottom: 0,
			} as DOMRect),
		}

		const hidePopup = () => {
			if (popupElement) {
				popupElement.style.display = 'none'
			}
		}

		const showPopup = () => {
			if (popupElement) {
				popupElement.style.display = 'block'
			}
		}

		const updatePosition = async () => {
			if (!popupElement) return
			await computePosition(virtualReference as any, popupElement, {
				placement: 'bottom-start',
				middleware: [
					offset(25),
					flip(),
					shift({padding: 8}),
				],
			}).then(({x, y}) => {
				if (popupElement) {
					popupElement.style.left = `${x}px`
					popupElement.style.top = `${y}px`
				}
			})
		}

		const updateAutocomplete = async (view: any, force = false) => {
			const {from} = view.state.selection
			const text = view.state.doc.textContent
			const textUpToCursor = text.substring(0, from)

			// Check if we're in an autocomplete context
			let autocompleteContext = null
			let fieldType: 'labels' | 'assignees' | 'projects' | null = null

			for (const field of AUTOCOMPLETE_FIELDS) {
				const pattern = new RegExp(`(${field}\\s*${FILTER_OPERATORS_REGEX}\\s*)(["']?)([^"'&|()]*)?$`, 'ig')
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
				hidePopup()
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
				hidePopup()
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
							const searchStart = currentText.lastIndexOf(item.context.search, from) + 1
							if (searchStart !== -1) {
								const transaction = view.state.tr.replaceWith(
									searchStart,
									searchStart + item.context.search.length,
									view.state.schema.text(newValue),
								)
								view.dispatch(transaction)
							}

							hidePopup()
						},
					},
					editor: this.editor,
				})
			} else {
				component.updateProps({
					items,
				})
			}

			// Create popup element on demand
			if (!popupElement) {
				popupElement = document.createElement('div')
				popupElement.style.position = 'absolute'
				popupElement.style.top = '0'
				popupElement.style.left = '0'
				popupElement.style.zIndex = '1000'
				popupElement.appendChild(component.element!)
				document.body.appendChild(popupElement)

				cleanupFloating = autoUpdate(virtualReference as any, popupElement, updatePosition)
			}

			// Update virtual reference to current cursor position
			const coords = view.coordsAtPos(from)
			const rect = {
				width: 0,
				height: 0,
				x: coords.left,
				y: coords.top,
				top: coords.top,
				left: coords.left,
				right: coords.left,
				bottom: coords.bottom,
			} as DOMRect
			virtualReference.getBoundingClientRect = () => rect

			showPopup()
			await updatePosition()
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
							if (cleanupFloating) {
								cleanupFloating()
							}
							if (popupElement && popupElement.parentNode) {
								popupElement.parentNode.removeChild(popupElement)
							}
							popupElement = null
							if (component) {
								component.destroy()
							}
						},
					}
				},
				props: {
					handleKeyDown(view, event) {
						if (!popupElement || popupElement.style.display === 'none') {
							return false
						}

						// Forward key events to the component
						if ((component as any)?.ref?.onKeyDown) {
							return (component as any).ref.onKeyDown({event})
						}

						return false
					},
				},
			}),
		]
	},
})
