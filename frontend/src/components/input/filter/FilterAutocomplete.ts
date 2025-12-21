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
	quoteChar: string // The quote character surrounding the keyword ('"', "'", or '' if unquoted)
}

interface SuggestionItem {
	id: number
	title?: string
	username?: string
	name?: string
}

export type AutocompleteField = 'labels' | 'assignees' | 'projects'

/**
 * Calculates the replacement range for autocomplete selection.
 * For single-value operators: replaces the entire keyword
 * For multi-value operators with commas: only replaces the text after the last comma
 * When inside quotes, extends the range to include the closing quote
 *
 * @param context - The autocomplete context containing position and keyword info
 * @param operator - The filter operator (e.g., 'in', '=', '?=')
 * @param hasClosingQuote - Whether there's a closing quote to include in replacement
 * @returns Object with replaceFrom and replaceTo positions
 */
export function calculateReplacementRange(
	context: { startPos: number; endPos: number; keyword: string },
	operator: string,
	hasClosingQuote: boolean = false,
): { replaceFrom: number; replaceTo: number } {
	// Add 1 to convert from string indices to ProseMirror positions
	// In ProseMirror, position 0 is before the document, text starts at position 1
	let replaceFrom = context.startPos + 1
	let replaceTo = context.endPos + 1

	// Handle multi-value operators - only replace the last value after comma
	if (isMultiValueOperator(operator) && context.keyword.includes(',')) {
		const lastCommaIndex = context.keyword.lastIndexOf(',')
		const textAfterComma = context.keyword.substring(lastCommaIndex + 1)
		const leadingSpaces = textAfterComma.length - textAfterComma.trimStart().length
		replaceFrom = context.startPos + lastCommaIndex + 1 + leadingSpaces + 1
	}

	// Extend range to include closing quote if present
	if (hasClosingQuote) {
		replaceTo += 1
	}

	return { replaceFrom, replaceTo }
}

export interface AutocompleteItem {
	id: number | string
	title: string
	item: ILabel | IUser | IProject
	fieldType: AutocompleteField
	context: AutocompleteContext
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

			// If at end of expression (nothing after), keep autocomplete open to allow selection
			if (trimmedAfter === '') {
				return false
			}

			// If there's a logical operator after, expression is complete (user has moved on)
			if (trimmedAfter.startsWith('&&') || trimmedAfter.startsWith('||') || trimmedAfter.startsWith(')')) {
				return true
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
					return labelStore.filterLabelsByQuery([], autocompleteContext.search).filter((label): label is ILabel => label !== undefined) as SuggestionItem[]
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
									// @ts-expect-error - projectId is used for URL replacement but not part of IAbstract
									assigneeSuggestions = await projectUserService.getAll({projectId: this.options.projectId}, {s: autocompleteContext.search}) as SuggestionItem[]
								} else {
									assigneeSuggestions = await userService.getAll({} as IUser, {s: autocompleteContext.search}) as SuggestionItem[]
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
					return projectStore.searchProject(autocompleteContext.search).filter((project): project is IProject => project !== undefined) as SuggestionItem[]
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

				if (match && match.index !== undefined) {
					const [, prefix = '', , quoteChar = '', keyword = ''] = match

					let search = keyword.trim()
					const operator = match[0].match(new RegExp(FILTER_OPERATORS_REGEX))?.[0] || ''
					if (operator === 'in' || operator === '?=') {
						const keywords = keyword.split(',')
						search = keywords[keywords.length - 1]?.trim() ?? ''
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
						quoteChar,
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
							const newValue = item.fieldType === 'assignees'
								? (item.item as IUser).username
								: (item.item as IProject | ILabel).title
							// Use currentAutocompleteContext (outer variable) for up-to-date positions
							// The local autocompleteContext would be stale since this callback
							// was created on first component render
							const context = currentAutocompleteContext
							if (!context) {
								return
							}
							const operator = context.operator

							// Check if there's a closing quote immediately after the keyword
							const docText = view.state.doc.textContent
							const charAfterKeyword = docText[context.endPos] || ''
							const hasClosingQuote = context.quoteChar !== '' && charAfterKeyword === context.quoteChar

							// Quote values that contain spaces for filter syntax
							// But skip quoting if already inside quotes (we'll replace including the closing quote)
							let insertValue: string = newValue ?? ''
							if (insertValue.includes(' ') && !context.quoteChar) {
								// Escape backslashes and quotes before wrapping in double quotes
								const escaped = insertValue.replace(/\\/g, '\\\\').replace(/"/g, '\\"')
								insertValue = `"${escaped}"`
							}
							const { replaceFrom, replaceTo } = calculateReplacementRange(context, operator, hasClosingQuote)

							const tr = view.state.tr.replaceWith(
								replaceFrom,
								replaceTo,
								view.state.schema.text(insertValue),
							)
							// Position cursor after the inserted text
							const newPos = replaceFrom + insertValue.length
							// @ts-expect-error - Selection.near is a static method but TypeScript doesn't recognize it on constructor
							tr.setSelection(view.state.selection.constructor.near(tr.doc.resolve(newPos)))
							view.dispatch(tr)
							
							// Update selection tracking
							lastSelectionPosition = newPos
							lastSelectionTime = Date.now()

							// Return focus to editor and position cursor
							setTimeout(() => {
								view.focus()
							}, 0)

							// Always suppress and hide after selection
							// User can type comma manually if they want to add more values
							suppressNextAutocomplete = true
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
