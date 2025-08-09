<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount} from 'vue'
import DatepickerWithValues from '@/components/date/DatepickerWithValues.vue'
import UserService from '@/services/user'
import AutocompleteDropdown from '@/components/input/AutocompleteDropdown.vue'
import {useLabelStore} from '@/stores/labels'
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'
import ProjectUserService from '@/services/projectUsers'
import {useProjectStore} from '@/stores/projects'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	AVAILABLE_FILTER_FIELDS,
	DATE_FIELDS,
	FILTER_JOIN_OPERATOR,
	FILTER_OPERATORS,
	FILTER_OPERATORS_REGEX,
	getFilterFieldRegexPattern,
	LABEL_FIELDS,
} from '@/helpers/filters'
import {useDebounceFn} from '@vueuse/core'
import {createRandomID} from '@/helpers/randomId'

// ProseMirror imports
import {EditorView, Decoration, DecorationSet} from '@tiptap/pm/view'
import {EditorState, Plugin, PluginKey} from '@tiptap/pm/state'
import {Schema} from '@tiptap/pm/model'
import {keymap} from '@tiptap/pm/keymap'
import {history, undo, redo} from '@tiptap/pm/history'
import {baseKeymap} from '@tiptap/pm/commands'

const props = defineProps<{
	modelValue: string,
	projectId?: number,
	inputLabel?: string,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'blur': [],
}>()

const userService = new UserService()
const projectUserService = new ProjectUserService()
const labelStore = useLabelStore()
const projectStore = useProjectStore()

const editorRef = ref<HTMLDivElement | null>(null)
const editor = ref<EditorView | null>(null)
const id = ref(createRandomID())


// Simple schema for plain text with highlighting
const filterSchema = new Schema({
	nodes: {
		doc: {
			content: 'paragraph*',
		},
		paragraph: {
			content: 'text*',
			group: 'block',
			parseDOM: [{tag: 'p'}],
			toDOM() { return ['p', 0] },
		},
		text: {
			group: 'inline',
		},
	},
	marks: {},
})

// Plugin for syntax highlighting
function createHighlightPlugin() {
	return new Plugin({
		key: new PluginKey('filterHighlight'),
		state: {
			init() {
				return DecorationSet.empty
			},
			apply(tr, decorationSet) {
				if (!tr.docChanged) {
					return decorationSet.map(tr.mapping, tr.doc)
				}
				return createDecorations(tr.doc)
			},
		},
		props: {
			decorations(state) {
				return this.getState(state)
			},
		},
	})
}

function createDecorations(doc: {textContent: string}) {
	const decorations: Decoration[] = []
	const text = doc.textContent

	// Helper function to add decoration
	const addDecoration = (from: number, to: number, className: string, attributes = {}) => {
		if (from < to && from >= 0 && to <= text.length) {
			decorations.push(
				Decoration.inline(from, to, {
					class: className,
					...attributes,
				}),
			)
		}
	}

	try {
		// Highlight filter fields
		AVAILABLE_FILTER_FIELDS.forEach(field => {
			const regex = new RegExp(`\\b${field}\\b`, 'gi')
			let match
			while ((match = regex.exec(text)) !== null) {
				addDecoration(match.index, match.index + match[0].length, 'filter-field')
			}
		})

		// Highlight operators
		FILTER_OPERATORS.forEach(op => {
			const escapedOp = op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
			const regex = new RegExp(`\\s(${escapedOp})\\s`, 'gi')
			let match
			while ((match = regex.exec(text)) !== null) {
				addDecoration(match.index + 1, match.index + 1 + match[1].length, 'filter-operator')
			}
		})

		// Highlight join operators
		FILTER_JOIN_OPERATOR.forEach(joinOp => {
			const regex = new RegExp(`\\b${joinOp}\\b`, 'gi')
			let match
			while ((match = regex.exec(text)) !== null) {
				addDecoration(match.index, match.index + match[0].length, 'filter-join-operator')
			}
		})

		// Highlight date values with clickable decoration
		DATE_FIELDS.forEach(dateField => {
			const pattern = new RegExp(dateField + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'gi')
			let match
			while ((match = pattern.exec(text)) !== null) {
				if (match[3]) { // If there's a value
					const valueStart = match.index + match[0].indexOf(match[3])
					const valueEnd = valueStart + match[3].length
					addDecoration(valueStart, valueEnd, 'filter-date-value', {
						'data-date-value': match[3],
						'data-position': valueStart.toString(),
					})
				}
			}
		})

		// Highlight assignee values
		ASSIGNEE_FIELDS.forEach(assigneeField => {
			const pattern = new RegExp(assigneeField + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'gi')
			let match
			while ((match = pattern.exec(text)) !== null) {
				if (match[3]) { // If there's a value
					const valueStart = match.index + match[0].indexOf(match[3])
					const valueEnd = valueStart + match[3].length
					addDecoration(valueStart, valueEnd, 'filter-assignee-value')
				}
			}
		})

		// Highlight label values with colors (simplified)
		LABEL_FIELDS.forEach(labelField => {
			const pattern = getFilterFieldRegexPattern(labelField)
			let match
			while ((match = pattern.exec(text)) !== null) {
				if (match[4]) { // If there's a value
					const valueStart = match.index + match[0].indexOf(match[4])
					const valueEnd = valueStart + match[4].length
					addDecoration(valueStart, valueEnd, 'filter-label-value')
				}
			}
		})
	} catch (error) {
		console.warn('Error creating decorations:', error)
	}

	return DecorationSet.create(doc, decorations)
}

// Initialize ProseMirror editor
onMounted(() => {
	if (!editorRef.value) return

	editor.value = new EditorView(editorRef.value, {
		state: createEditorState(props.modelValue),
		dispatchTransaction(tr) {
			if (!editor.value) return
			
			const newState = editor.value.state.apply(tr)
			editor.value.updateState(newState)
			
			// Update the model value when document changes
			if (tr.docChanged) {
				const text = newState.doc.textContent
				emit('update:modelValue', text)
			}
		},
		attributes: {
			class: 'filter-prosemirror',
			style: 'white-space: pre-wrap',
		},
		handleDOMEvents: {
			click(_, event) {
				const target = event.target as HTMLElement
				if (target.classList.contains('filter-date-value')) {
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
			input() {
				handleFieldInput()
				return false
			},
		},
	})

})

onBeforeUnmount(() => {
	editor.value?.destroy()
})

// Create a new editor state similar to the working draft approach
function createEditorState(content = '') {
	const nodes = content ? [
		filterSchema.node('paragraph', null, [
			filterSchema.text(content),
		]),
	] : [filterSchema.node('paragraph')]

	return EditorState.create({
		schema: filterSchema,
		plugins: [
			keymap({
				...baseKeymap,
				'Mod-z': undo,
				'Mod-y': redo,
				'Enter': () => {
					blurDebounced()
					return true
				},
			}),
			history(),
			createHighlightPlugin(),
		],
		doc: filterSchema.node('doc', null, nodes),
	})
}


const currentOldDatepickerValue = ref('')
const currentDatepickerValue = ref('')
const currentDatepickerPos = ref(0)
const datePickerPopupOpen = ref(false)

function updateDateInQuery(newDate: string | Date | null) {
	if (!editor.value || !newDate) return
	
	const dateStr = typeof newDate === 'string' ? newDate : newDate.toISOString().split('T')[0]
	const currentText = editor.value.state.doc.textContent
	const newText = currentText.replace(currentOldDatepickerValue.value, dateStr)
	currentOldDatepickerValue.value = dateStr
	
	// Update by recreating the editor state
	const newState = createEditorState(newText)
	editor.value.updateState(newState)
	emit('update:modelValue', newText)
}

const autocompleteMatchPosition = ref(0)
const autocompleteMatchText = ref('')
const autocompleteResultType = ref<'labels' | 'assignees' | 'projects' | null>(null)
const autocompleteResults = ref<Array<{id: number, title?: string, username?: string, name?: string}>>([])

function handleFieldInput() {
	if (!editor.value) return
	
	const state = editor.value.state
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
		if (matched.startsWith('label')) {
			autocompleteResultType.value = 'labels'
			autocompleteResults.value = labelStore.filterLabelsByQuery([], search)
		}
		if (matched.startsWith('assignee')) {
			autocompleteResultType.value = 'assignees'
			if (props.projectId) {
				projectUserService.getAll({projectId: props.projectId}, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			} else {
				userService.getAll({}, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			}
		}
		if (!props.projectId && matched.startsWith('project')) {
			autocompleteResultType.value = 'projects'
			autocompleteResults.value = projectStore.searchProject(search)
		}
		autocompleteMatchText.value = keyword
		autocompleteMatchPosition.value = match.index + prefix.length - 1 + keyword.replace(search, '').length
	})
}

function autocompleteSelect(value: {id: number, username?: string, title?: string}) {
	if (!editor.value) return
	
	const newValue = autocompleteResultType.value === 'assignees' ? value.username : value.title
	const currentText = editor.value.state.doc.textContent
	const newText = currentText.substring(0, autocompleteMatchPosition.value + 1) +
		newValue +
		currentText.substring(autocompleteMatchPosition.value + autocompleteMatchText.value.length + 1)
	
	// Update by recreating the editor state
	const newState = createEditorState(newText)
	editor.value.updateState(newState)
	emit('update:modelValue', newText)
	
	autocompleteResults.value = []
}

// The blur from the editor might happen before the replacement after autocomplete select was done.
const blurDebounced = useDebounceFn(() => emit('blur'), 500)
</script>

<template>
	<div class="field">
		<label
			class="label"
			:for="id"
		>
			{{ inputLabel ?? $t('filters.query.title') }}
		</label>
		<AutocompleteDropdown
			:options="autocompleteResults"
			@blur="editor?.dom.blur()"
			@update:modelValue="autocompleteSelect"
		>
			<template
				#input="{ onKeydown, onFocusField }"
			>
				<div class="control filter-input">
					<div
						:id="id"
						ref="editorRef"
						class="filter-editor-container"
						:class="{'has-autocomplete-results': autocompleteResults.length > 0}"
						@focus="onFocusField"
						@keydown="onKeydown"
						@blur="blurDebounced"
					/>
					<DatepickerWithValues
						v-model="currentDatepickerValue"
						v-model:open="datePickerPopupOpen"
						@update:modelValue="updateDateInQuery"
					/>
				</div>
			</template>
			<template
				#result="{ item }"
			>
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
	</div>
</template>

<style lang="scss">
.filter-editor-container {
	min-block-size: 2.5em;
	border: 1px solid var(--input-border-color);
	border-radius: var(--input-radius);
	padding: .5em .75em;
	background: var(--white);
	position: relative;

	&.has-autocomplete-results {
		border-radius: var(--input-radius) var(--input-radius) 0 0;
	}

	&:focus-within {
		border-color: var(--primary);
		box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.25);
	}

	.filter-prosemirror {
		outline: none;
		min-block-size: 1.5em;
		line-height: 1.5;

		// Placeholder support
		&:empty::before {
			content: attr(data-placeholder);
			color: var(--input-placeholder-color);
			pointer-events: none;
			position: absolute;
		}

		// Syntax highlighting styles
		.filter-field {
			color: var(--code-literal);
			font-weight: 600;
		}

		.filter-operator {
			color: var(--code-keyword);
			font-weight: 600;
		}

		.filter-join-operator {
			color: var(--code-section);
			font-weight: 600;
		}

		.filter-date-value {
			background-color: var(--primary);
			color: var(--white);
			border-radius: var(--radius);
			padding: 0.125em 0.25em;
			cursor: pointer;
			transition: background-color var(--transition);

			&:hover {
				background-color: var(--primary-dark);
			}
		}

		.filter-label-value {
			border-radius: var(--radius);
			background-color: var(--grey-200);
			color: var(--grey-700);
			padding: 0.125em 0.25em;
		}

		.filter-assignee-value {
			border-radius: var(--radius);
			background-color: var(--grey-200);
			color: var(--grey-700);
			padding: 0.125em 0.25em;
		}
	}

	// ProseMirror base styles
	.ProseMirror {
		outline: none;
		min-block-size: 1.5em;

		p {
			margin: 0;
		}
	}
}
</style>

<style lang="scss" scoped>
.filter-input {
	position: relative;
}
</style>
