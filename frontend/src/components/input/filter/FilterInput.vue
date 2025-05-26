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

const emit = defineEmits(['update:filter'])
const editorRef = ref<HTMLDivElement | null>(null)
let editorView: EditorView | null = null

const {t} = useI18n()

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
			}
		},
		dispatchTransaction(transaction) {
			if (!editorView) return

			const newState = editorView.state.apply(transaction)
			editorView.updateState(newState)

			// When the document changes, emit the updated filter value
			if (transaction.docChanged) {
				const snakeCaseFilter = processContent(editorView)
				emit('update:filter', snakeCaseFilter)
				filterValue.value = snakeCaseFilter
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
</script>

<template>
	<div class="filter-input">
		<div ref="editorRef" class="editor-content"></div>
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
	padding: .5rem .75rem;
	background: var(--white);
	overflow: hidden;
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
