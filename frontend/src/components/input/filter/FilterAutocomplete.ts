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

interface AutocompleteContext {
		field: string
		prefix: string
		keyword: string
		search: string
		operator: string
		startPos: number
		endPos: number
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
		let currentAutocompleteContext: AutocompleteContext | null = null
		let cleanupFloating: (() => void) | null = null
		let suppressNextAutocomplete = false
		let clickOutsideHandler: ((event: MouseEvent) => void) | null = null
		let debounceTimer: NodeJS.Timeout | null = null

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
			currentAutocompleteContext = null
			
			// Remove click outside handler
			if (clickOutsideHandler) {
				document.removeEventListener('mousedown', clickOutsideHandler)
				clickOutsideHandler = null
			}
			
			// Clear component items to avoid stale state
			if (component) {
				component.updateProps({
					items: [],
				})
			}
		}

		const showPopup = () => {
			if (popupElement) {
				popupElement.style.display = 'block'
				
				// Add click outside handler if not already added
				if (!clickOutsideHandler) {
					clickOutsideHandler = (event: MouseEvent) => {
						const target = event.target as Node
						const editorElement = (this.editor?.view?.dom) as Node
						
						// Don't hide if clicking inside the popup or the editor
						if (popupElement?.contains(target) || editorElement?.contains(target)) {
							return
						}
						
						hidePopup()
					}
					document.addEventListener('mousedown', clickOutsideHandler)
				}
			}
		}

		const fetchSuggestions = async (autocompleteContext: AutocompleteContext, fieldType: 'labels' | 'assignees' | 'projects') => {
			let suggestions: Array<{id: number, title?: string, username?: string, name?: string}> = []

			try {
				if (fieldType === 'labels') {
					// Local search, no debouncing needed
					suggestions = labelStore.filterLabelsByQuery([], autocompleteContext.search)
				} else if (fieldType === 'assignees') {

					if (debounceTimer) {
						clearTimeout(debounceTimer)
					}

					return new Promise((resolve) => {
						debounceTimer = setTimeout(async () => {
							let assigneeSuggestions: Array<{id: number, username: string, name?: string}> = []
							try {
								if (this.options.projectId) {
									assigneeSuggestions = await projectUserService.getAll({projectId: this.options.projectId}, {s: autocompleteContext.search})
								} else {
									assigneeSuggestions = await userService.getAll({}, {s: autocompleteContext.search})
								}
								// For assignees, show suggestions even with empty search, but limit if we have many
								if (autocompleteContext.search === '' && assigneeSuggestions.length > 10) {
									assigneeSuggestions = assigneeSuggestions.slice(0, 10)
								}
							} catch (error) {
								console.error('Error fetching assignee suggestions:', error)
								assigneeSuggestions = []
							}
							resolve(assigneeSuggestions)
						}, 300) // 300ms debounce delay
					})
				} else if (fieldType === 'projects' && !this.options.projectId) {
					// Local search, no debouncing needed
					suggestions = projectStore.searchProject(autocompleteContext.search)
				}
			} catch (error) {
				console.error('Error fetching suggestions:', error)
				suggestions = []
			}

			return suggestions
		}

		const updatePosition = async () => {
			if (!popupElement) return
			await computePosition(virtualReference, popupElement, {
				placement: 'bottom-start',
				strategy: 'fixed',
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

		const updateAutocomplete = async (view: {state: {selection: {from: number}, doc: {textContent: string}}, coordsAtPos: (pos: number) => {left: number, top: number, bottom: number}, dispatch: (tr: unknown) => void, focus: () => void}, force = false) => {
			if (suppressNextAutocomplete) {
				suppressNextAutocomplete = false
				hidePopup()
				return
			}
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
						operator,
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

			// Get suggestions based on field type (debounced for API calls only)
			const suggestions = await fetchSuggestions(autocompleteContext, fieldType)

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
						command: (item: {id: number, title: string, description: string, item: {id: number, title?: string, username?: string, name?: string}, fieldType: string, context: {field: string, prefix: string, keyword: string, search: string, operator: string, startPos: number, endPos: number}}) => {
							// Handle selection
							const newValue = item.fieldType === 'assignees' ? item.item.username : item.item.title
							const {from} = view.state.selection
							const context = item.context
							const operator = context.operator
							
							let insertValue = newValue
							const replaceFrom = Math.max(0, from - context.search.length)
							const replaceTo = from
							
							// Handle multi-value operators
							const isMultiValueOperator = operator === 'in' || operator === '?=' || operator === 'not in' || operator === '?!='
							if (isMultiValueOperator && context.keyword.includes(',')) {
								// For multi-value fields, we need to replace only the current search term
								const keywords = context.keyword.split(',')
								const currentKeywordIndex = keywords.length - 1
								
								// If we're not adding the first item, add comma prefix
								if (currentKeywordIndex > 0 && keywords[currentKeywordIndex].trim() === context.search.trim()) {
									// We're replacing the last incomplete keyword
									insertValue = newValue
								} else {
									// We're adding to existing keywords
									insertValue = ',' + newValue
								}
							}
							
							const tr = view.state.tr.replaceWith(
								replaceFrom,
								replaceTo,
								view.state.schema.text(insertValue),
							)
							// Position cursor after the inserted text
							const newPos = replaceFrom + insertValue.length
							tr.setSelection(view.state.selection.constructor.near(tr.doc.resolve(newPos)))
							view.dispatch(tr)

							// Return focus to editor and position cursor
							setTimeout(() => {
								view.focus()
							}, 0)

							// For multi-value operators, don't suppress autocomplete to keep dropdown open
							if (isMultiValueOperator) {
								// Add comma and space for next entry if not already present
								setTimeout(() => {
									const currentText = view.state.doc.textContent
									const currentPos = view.state.selection.from
									if (currentText.charAt(currentPos) !== ',') {
										const tr = view.state.tr.insertText(',', currentPos)
										view.dispatch(tr)
									}
								}, 10)
							} else {
								suppressNextAutocomplete = true
								hidePopup()
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

			// Create popup element on demand
			if (!popupElement) {
				popupElement = document.createElement('div')
				popupElement.style.position = 'fixed'
				popupElement.style.top = '0'
				popupElement.style.left = '0'
				popupElement.style.zIndex = '20000'
				popupElement.id = 'filter-autocomplete-popup'
				popupElement.appendChild(component.element!)
				document.body.appendChild(popupElement)

				cleanupFloating = autoUpdate(virtualReference, popupElement, updatePosition)
			}

			// Update virtual reference to start of the current search token
			const anchorFrom = autocompleteContext ? Math.max(0, from - (autocompleteContext.search?.length || 0)) : from
			const coords = view.coordsAtPos(anchorFrom)
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
							if (clickOutsideHandler) {
								document.removeEventListener('mousedown', clickOutsideHandler)
								clickOutsideHandler = null
							}
							if (debounceTimer) {
								clearTimeout(debounceTimer)
								debounceTimer = null
							}
							if (popupElement && popupElement.parentNode) {
								popupElement.parentNode.removeChild(popupElement)
							}
							popupElement = null
							suppressNextAutocomplete = false
							currentAutocompleteContext = null
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
						if ((component as VueRenderer & {ref?: {onKeyDown?: (params: {event: KeyboardEvent}) => boolean}})?.ref?.onKeyDown) {
							return (component as VueRenderer & {ref: {onKeyDown: (params: {event: KeyboardEvent}) => boolean}}).ref.onKeyDown({event})
						}

						return false
					},
				},
			}),
		]
	},
})
