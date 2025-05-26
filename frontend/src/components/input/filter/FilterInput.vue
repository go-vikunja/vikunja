<script setup lang="ts">
import {defineEmits, onMounted, ref} from 'vue'
import {EditorState} from '@tiptap/pm/state'
import {EditorView} from '@tiptap/pm/view'
import {keymap} from '@tiptap/pm/keymap'
import {baseKeymap} from '@tiptap/pm/commands'

import {filterHighlighter} from './highlighter.ts'
import {schema} from './schema.ts'

const emit = defineEmits(['update:filter'])
const editorRef = ref<HTMLDivElement | null>(null)
let editorView: EditorView | null = null

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
</script>

<template>
	<div class="filter-input">
		<div ref="editorRef" class="editor-content"></div>
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
}

.editor-content {
	line-height: 1.5;
}

.ProseMirror {
	outline: none;
	white-space: pre-wrap;

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
}
</style>
