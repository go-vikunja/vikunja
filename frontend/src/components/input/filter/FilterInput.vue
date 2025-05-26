<script setup lang="ts">
import {defineEmits, onMounted, ref} from 'vue'
import {EditorState} from '@tiptap/pm/state'
import {EditorView} from '@tiptap/pm/view'
import {keymap} from '@tiptap/pm/keymap'
import {baseKeymap} from '@tiptap/pm/commands'

import {filterHighlighter} from './highlighter.ts'
import {schema} from './schema.ts'
import {placeholder} from '@/components/input/filter/placeholder.ts'
import {useI18n} from 'vue-i18n'
import DatepickerWithValues from '@/components/date/DatepickerWithValues.vue'
import AutocompleteDropdown from '@/components/input/AutocompleteDropdown.vue'
import UserService from '@/services/user'
import ProjectUserService from '@/services/projectUsers'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	FILTER_OPERATORS_REGEX,
	LABEL_FIELDS,
	PROJECT_FIELDS,
} from '@/helpers/filters'
import {useDebounceFn} from '@vueuse/core'

const props = defineProps<{
	projectId?: number,
}>()

const emit = defineEmits(['update:filter'])
const editorRef = ref<HTMLDivElement | null>(null)
let editorView: EditorView | null = null

const {t} = useI18n()

// Services and stores for autocomplete
const userService = new UserService()
const projectUserService = new ProjectUserService()
const labelStore = useLabelStore()
const projectStore = useProjectStore()

// Autocomplete state
const autocompleteMatchPosition = ref(0)
const autocompleteMatchText = ref('')
const autocompleteResultType = ref<'labels' | 'assignees' | 'projects' | null>(null)
const autocompleteResults = ref<any[]>([])

// Store references to the dropdown functions
const dropdownOnFocusField = ref<(() => void) | null>(null)
const dropdownOnKeydown = ref<((event: KeyboardEvent) => void) | null>(null)

// Set up the editor state with our custom schema
const createEditorState = (content = '') => {
	const nodes = content ? [
		schema.node('paragraph', null, [
			schema.text(content),
		]),
	] : [schema.node('paragraph')]

	return EditorState.create({
		schema: schema,
		plugins: [
			keymap(baseKeymap),
			filterHighlighter,
			placeholder(t('filters.query.placeholder')),
		],
		doc: schema.node('doc', null, nodes),
	})
}

// Process the editor content to output snake_cased filter
const processContent = (view: EditorView) => {
	if (!view) return ''

	const content = view.state.doc.textContent

	const fieldMap: Record<string, string> = {
		'dueDate': 'due_date',
		'percentDone': 'percent_done',
		'startDate': 'start_date',
		'endDate': 'end_date',
		'doneAt': 'done_at',
	}

	// Simple regex-based transformation for field names
	let processed = content
	Object.entries(fieldMap).forEach(([camel, snake]) => {
		const regex = new RegExp(`\\b${camel}\\b`, 'g')
		processed = processed.replace(regex, snake)
	})

	return processed
}

// Initialize the editor when the component is mounted
onMounted(() => {
	if (!editorRef.value) return

	editorView = new EditorView(editorRef.value, {
		state: createEditorState(),
		attributes: {
			spellcheck: 'false',
		},
		handleDOMEvents: {
			click(view, event) {
				const target = event.target as HTMLElement
				if (target.classList.contains('date-value')) {
					event.preventDefault()
					event.stopPropagation()
					
					const dateValue = target.getAttribute('data-date-value') || ''
					const position = parseInt(target.getAttribute('data-position') || '0')
					
					currentOldDatepickerValue.value = dateValue
					currentDatepickerValue.value = dateValue
					currentDatepickerPos.value = position
					datePickerPopupOpen.value = true
					
					return true
				}
				return false
			},
			focus(view, event) {
				if (dropdownOnFocusField.value) {
					dropdownOnFocusField.value()
				}
				return false
			},
			keydown(view, event) {
				if (dropdownOnKeydown.value) {
					dropdownOnKeydown.value(event as KeyboardEvent)
				}
				return false
			}
		},
		dispatchTransaction(transaction) {
			if (!editorView) return

			const newState = editorView.state.apply(transaction)
			editorView.updateState(newState)

			// When the document changes, emit the updated filter value and handle autocomplete
			if (transaction.docChanged) {
				const snakeCaseFilter = processContent(editorView)
				emit('update:filter', snakeCaseFilter)
				filterValue.value = snakeCaseFilter
				
				// Handle autocomplete with the updated content
				handleFieldInput()
			}
		},
	})
})

const filterValue = ref('')

// Date picker functionality
const currentOldDatepickerValue = ref('')
const currentDatepickerValue = ref('')
const currentDatepickerPos = ref(0)
const datePickerPopupOpen = ref(false)

function updateDateInQuery(newDate: string | Date | null) {
	if (!editorView || !newDate) return
	
	const dateStr = typeof newDate === 'string' ? newDate : newDate.toISOString().split('T')[0]
	const currentText = editorView.state.doc.textContent
	const newText = currentText.replace(currentOldDatepickerValue.value, dateStr)
	currentOldDatepickerValue.value = dateStr
	
	// Update by creating a transaction instead of recreating the state
	const tr = editorView.state.tr.replaceWith(0, editorView.state.doc.content.size, 
		editorView.state.schema.text(newText))
	editorView.dispatch(tr)
	
	emit('update:filter', processContent(editorView))
	filterValue.value = processContent(editorView)
}

// Autocomplete functionality
function handleFieldInput() {
	if (!editorView) return
	
	const state = editorView.state
	const selection = state.selection
	const cursorPosition = selection.from
	const text = state.doc.textContent
	const textUpToCursor = text.substring(0, cursorPosition)
	autocompleteResults.value = []

	AUTOCOMPLETE_FIELDS.forEach(field => {
		const pattern = new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()]+\\1?)?$', 'ig')
		const match = pattern.exec(textUpToCursor)

		if (match === null) {
			return
		}

		const [matched, prefix, operator, , keyword] = match
		if(!keyword) {
			return
		}

		let search = keyword
		if (operator === 'in' || operator === '?=') {
			const keywords = keyword.split(',')
			search = keywords[keywords.length - 1].trim()
		}
		
		if (LABEL_FIELDS.includes(field)) {
			autocompleteResultType.value = 'labels'
			autocompleteResults.value = labelStore.filterLabelsByQuery([], search)
		}
		if (ASSIGNEE_FIELDS.includes(field)) {
			autocompleteResultType.value = 'assignees'
			if (props.projectId) {
				projectUserService.getAll({projectId: props.projectId} as any, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			} else {
				userService.getAll({} as any, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			}
		}
		if (!props.projectId && PROJECT_FIELDS.includes(field)) {
			autocompleteResultType.value = 'projects'
			autocompleteResults.value = projectStore.searchProject(search)
		}
		autocompleteMatchText.value = keyword
		autocompleteMatchPosition.value = match.index + prefix.length - 1 + keyword.replace(search, '').length
	})
}

function autocompleteSelect(value: any) {
	if (!editorView) return
	
	const newValue = autocompleteResultType.value === 'assignees' ? value.username : value.title
	const currentText = editorView.state.doc.textContent
	const newText = currentText.substring(0, autocompleteMatchPosition.value + 1) +
		newValue +
		currentText.substring(autocompleteMatchPosition.value + autocompleteMatchText.value.length + 1)
	
	// Update by creating a transaction instead of recreating the state
	const tr = editorView.state.tr.replaceWith(0, editorView.state.doc.content.size, 
		editorView.state.schema.text(newText))
	editorView.dispatch(tr)
	
	emit('update:filter', processContent(editorView))
	filterValue.value = processContent(editorView)
	
	autocompleteResults.value = []
}

// The blur from the editor might happen before the replacement after autocomplete select was done.
const blurDebounced = useDebounceFn(() => {}, 500)

// Function to setup dropdown callbacks from the slot props
function setupDropdownCallbacks(onFocusField: () => void, onKeydown: (event: KeyboardEvent) => void) {
	dropdownOnFocusField.value = onFocusField
	dropdownOnKeydown.value = onKeydown
	return ''
}
</script>

<template>
	<div class="filter-input">
		<AutocompleteDropdown
			:options="autocompleteResults"
			@blur="editorRef?.blur()"
			@update:modelValue="autocompleteSelect"
		>
			<template #input="{ onKeydown, onFocusField }">
				<div 
					ref="editorRef" 
					class="editor-content"
					:class="{'has-autocomplete-results': autocompleteResults.length > 0}"
					@blur="blurDebounced"
				></div>
				<span v-show="false">{{ setupDropdownCallbacks(onFocusField, onKeydown) }}</span>
			</template>
			<template #result="{ item }">
				<XLabel
					v-if="autocompleteResultType === 'labels'"
					:label="item"
				/>
				<User
					v-else-if="autocompleteResultType === 'assignees'"
					:user="item"
					:avatar-size="25"
				/>
				<template v-else>
					{{ item.title }}
				</template>
			</template>
		</AutocompleteDropdown>
		<DatepickerWithValues
			class="filter-datepicker"
			v-model="currentDatepickerValue"
			v-model:open="datePickerPopupOpen"
			@update:modelValue="updateDateInQuery"
		/>
	</div>
	<pre>{{ filterValue }}</pre>
</template>

<style lang="scss">
.filter-input {
	border: 1px solid var(--input-border-color);
	border-radius: var(--input-radius);
	background: var(--white);
	transition: border-color 0.2s ease;

	&:focus-within {
		border-color: var(--primary);
	}
	
	.filter-datepicker {
		position: absolute;
	}
}

.editor-content {
	line-height: 1.5;
	padding: .5rem .75rem;
}

.ProseMirror {
	outline: none;
	white-space: pre-wrap;
	padding: 0 !important;

	.field {
		color: var(--code-literal);
	}

	.operator {
		color: var(--code-keyword);
	}

	.value {
		border-radius: $radius;
		padding: .125rem .25rem;
		background: var(--grey-100);
	}

	.label-value {
		border-radius: $radius;
		padding: .125rem .25rem;
		font-weight: 500;
	}

	.date-value {
		background-color: var(--primary);
		color: var(--white);
		border-radius: $radius;
		padding: 0.125em 0.25em;
		cursor: pointer;
		transition: background-color var(--transition);

		&:hover {
			background-color: var(--primary-dark);
		}
	}

	.grouping, .logical {
		color: var(--code-section);
	}

	.user-value {
		position: relative;
		padding-left: 1.5em;

		&::before {
			content: attr(data-user);
			position: absolute;
			left: 0;
			top: 50%;
			transform: translateY(-50%);
			width: 1.2em;
			height: 1.2em;
			background-color: #3b82f6;
			color: white;
			border-radius: 50%;
			font-size: 0.8em;
			display: flex;
			align-items: center;
			justify-content: center;
			text-transform: uppercase;
		}
	}

	&[data-placeholder]::before {
		color: var(--grey-500);
		position: absolute;
		content: attr(data-placeholder);
		pointer-events: none;
	}
}
</style>
