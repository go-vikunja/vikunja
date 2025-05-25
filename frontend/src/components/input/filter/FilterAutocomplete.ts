import {Extension} from '@tiptap/core'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {VueRenderer} from '@tiptap/vue-3'
import type { EditorView } from '@tiptap/pm/view'
import {computePosition, flip, shift, offset, autoUpdate} from '@floating-ui/dom'

import FilterCommandsList from './FilterCommandsList.vue'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	FILTER_OPERATORS_REGEX,
	isMultiValueOperator,
	LABEL_FIELDS,
	PROJECT_FIELDS,
} from '@/helpers/filters'

import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import UserService from '@/services/user'
import ProjectUserService from '@/services/projectUsers'
import type { IUser } from '@/modelTypes/IUser'
import type { IProject } from '@/modelTypes/IProject'
import type { ILabel } from '@/modelTypes/ILabel'

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
	isComplete: boolean
}

interface SuggestionItem {
	id: number
	title?: string
	username?: string
	name?: string
}

export type AutocompleteField = 'labels' | 'assignees' | 'projects'

export interface AutocompleteItem {
	id: number | string
	title: string
	item: ILabel | IUser | IProject
	fieldType: AutocompleteField
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
		let lastSelectionPosition = -1
		let lastSelectionTime = 0

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

		const isFilterExpressionComplete = (textAfterExpression: string, keyword: string, operator: string): boolean => {
			// If the keyword is empty, it's definitely not complete
			if (!keyword.trim()) {
				return false
			}

			// Check if we're immediately after a recent selection
			const timeSinceLastSelection = Date.now() - lastSelectionTime
			if (timeSinceLastSelection < 1000) { // 1 second grace period
				return true
			}

			// For multi-value operators, check if we're in the middle of typing multiple values
			if (isMultiValueOperator(operator) && keyword.includes(',')) {
				const lastValue = keyword.split(',').pop()?.trim() || ''
				// If the last value after comma is empty or very short, we're likely still typing
				return lastValue.length > 1
			}

			// Check what comes after the expression
			const trimmedAfter = textAfterExpression.trim()
			
			// If there's a logical operator or end of string immediately after, it's likely complete
			if (trimmedAfter === '' || trimmedAfter.startsWith('&&') || trimmedAfter.startsWith('||') || trimmedAfter.startsWith(')')) {
				return keyword.trim().length > 1
			}

			// If there's a space followed by non-operator text, it's likely complete
			if (trimmedAfter.startsWith(' ') && !trimmedAfter.match(/^\s*[&|()]/)) {
				return true
			}

			return false
		}

		const hidePopup = () => {
			if (popupElement) {
				popupElement.style.display = 'none'
			}
			currentAutocompleteContext = null
			
			if (clickOutsideHandler) {
				document.removeEventListener('mousedown', clickOutsideHandler)
				clickOutsideHandler = null
			}
			
			if (component) {
				component.updateProps({
					items: [],
				})
			}
		}

		const showPopup = () => {
			if (popupElement) {
				popupElement.style.display = 'block'
				
				if (!clickOutsideHandler) {
					clickOutsideHandler = (event: MouseEvent) => {
						const target = event.target as Node
						const editorElement = (this.editor?.view?.dom) as Node
						
						if (popupElement?.contains(target) || editorElement?.contains(target)) {
							return
						}
						
						hidePopup()
					}
					document.addEventListener('mousedown', clickOutsideHandler)
				}
			}
		}

		const fetchSuggestions = async (autocompleteContext: AutocompleteContext, fieldType: AutocompleteField): Promise<SuggestionItem[]> => {
			try {
				if (fieldType === 'labels') {
					return labelStore.filterLabelsByQuery([], autocompleteContext.search)
				}

				if (fieldType === 'assignees') {

					if (debounceTimer) {
						clearTimeout(debounceTimer)
					}

					return new Promise((resolve) => {
						debounceTimer = setTimeout(async () => {
							let assigneeSuggestions: SuggestionItem[] = []
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
						}, 300)
					})
				}
				
				if (fieldType === 'projects' && !this.options.projectId) {
					return projectStore.searchProject(autocompleteContext.search)
				}
			} catch (error) {
				console.error('Error fetching suggestions:', error)
				return []
			}
			
			console.error('Unknown field type:', fieldType)

			return []
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

		const updateAutocomplete = async (view: EditorView, force: boolean = false) => {
			const {from} = view.state.selection

			if (suppressNextAutocomplete) {
				suppressNextAutocomplete = false
				hidePopup()
				return
			}

			// Check if we're too close to a recent selection (position-based suppression)
			if (lastSelectionPosition >= 0 && Math.abs(from - lastSelectionPosition) <= 2) {
				const timeSinceLastSelection = Date.now() - lastSelectionTime
				if (timeSinceLastSelection < 500) {
					hidePopup()
					return
				}
			}

			const text = view.state.doc.textContent
			const textUpToCursor = text.substring(0, from)

			let autocompleteContext: AutocompleteContext | null = null
			let fieldType: AutocompleteField | null = null

			for (const field of AUTOCOMPLETE_FIELDS) {
				const pattern = new RegExp(`(${field}\\s*${FILTER_OPERATORS_REGEX}\\s*)(["']?)([^"'&|()]*)?$`, 'ig')
				const match = pattern.exec(textUpToCursor)

				if (match) {
					const [, prefix, , , keyword = ''] = match

					let search = keyword.trim()
					const operator = match[0].match(new RegExp(FILTER_OPERATORS_REGEX))?.[0] || ''
					if (operator === 'in' || operator === '?=') {
						const keywords = keyword.split(',')
						search = keywords[keywords.length - 1].trim()
					}

					// Check if this expression is complete
					const textAfterExpression = text.substring(from)
					const isComplete = isFilterExpressionComplete(textAfterExpression, keyword, operator)

					autocompleteContext = {
						field,
						prefix,
						keyword,
						search,
						operator,
						startPos: match.index + prefix.length,
						endPos: match.index + prefix.length + keyword.length,
						isComplete,
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

			if (!autocompleteContext || !fieldType || autocompleteContext.isComplete) {
				hidePopup()
				return
			}

			const suggestions = await fetchSuggestions(autocompleteContext, fieldType)

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

			if (!component) {
				component = new VueRenderer(FilterCommandsList, {
					props: {
						items,
						command: (item: AutocompleteItem) => {
							// Handle selection
							const newValue = item.fieldType === 'assignees' ? item.item.username : item.item.title
							const {from} = view.state.selection
							const context = autocompleteContext
							const operator = context.operator
							
							let insertValue: string = newValue ?? ''
							const replaceFrom = Math.max(0, from - context.search.length)
							const replaceTo = from
							
							// Handle multi-value operators
							if (isMultiValueOperator(operator) && context.keyword.includes(',')) {
								// For multi-value fields, we need to replace only the current search term
								const keywords = context.keyword.split(',')
								const currentKeywordIndex = keywords.length - 1
								
								// If we're not adding the first item, add comma prefix
								if (currentKeywordIndex > 0 && keywords[currentKeywordIndex].trim() === context.search.trim()) {
									// We're replacing the last incomplete keyword
									insertValue = newValue ?? ''
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
							
							// Update selection tracking
							lastSelectionPosition = newPos
							lastSelectionTime = Date.now()

							// Return focus to editor and position cursor
							setTimeout(() => {
								view.focus()
							}, 0)

							// For multi-value operators, don't suppress autocomplete to keep dropdown open
							if (isMultiValueOperator(operator)) {
								// Add comma and space for next entry if not already present
								setTimeout(() => {
									const currentText = view.state.doc.textContent
									const currentPos = view.state.selection.from
									if (currentText.charAt(currentPos) !== ',') {
										const tr = view.state.tr.insertText(',', currentPos)
										view.dispatch(tr)
										// Update position after comma insertion
										lastSelectionPosition = currentPos + 1
										lastSelectionTime = Date.now()
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
							lastSelectionPosition = -1
							lastSelectionTime = 0
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
