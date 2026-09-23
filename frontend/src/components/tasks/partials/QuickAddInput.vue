<script setup lang="ts">
import {onBeforeUnmount, watch} from 'vue'
import {EditorContent, useEditor} from '@tiptap/vue-3'
import {Extension, type JSONContent} from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import {Placeholder} from '@tiptap/extensions'

import {useLabels} from '@/composables/useLabels'
import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {isAutoFocusViewport} from '@/directives/focus'
import {createQuickAddHighlighter, quickAddHighlighterKey} from './quickAddHighlighter'

const props = defineProps<{
	placeholder: string,
	autofocus?: boolean,
}>()

const emit = defineEmits<{
	submit: [],
}>()

const modelValue = defineModel<string>({required: true})

const authStore = useAuthStore()
const {labels} = useLabels()
const projectList = useProjects()

function projectExists(title: string) {
	const lower = title.toLowerCase()
	return projectList.findProjectByExactname(title) !== undefined
		|| Object.values(projectList.projects).some(p => p.identifier?.toLowerCase() === lower)
}

// Text nodes can't be empty, so blank lines become empty paragraphs.
function toDoc(text: string): JSONContent {
	return {
		type: 'doc',
		content: text.split('\n').map(line => ({
			type: 'paragraph',
			content: line === '' ? [] : [{type: 'text', text: line}],
		})),
	}
}

const editor = useEditor({
	editorProps: {
		attributes: () => ({
			role: 'textbox',
			'aria-multiline': 'true',
			'aria-label': props.placeholder,
			class: 'add-task-textarea',
		}),
	},
	extensions: [
		// Plain text only: marks and list input rules would eat magic like "*label" or "- subtask".
		StarterKit.configure({
			blockquote: false,
			bold: false,
			bulletList: false,
			code: false,
			codeBlock: false,
			dropcursor: false,
			gapcursor: false,
			hardBreak: false,
			heading: false,
			horizontalRule: false,
			italic: false,
			link: false,
			listItem: false,
			listKeymap: false,
			orderedList: false,
			strike: false,
			trailingNode: false,
			underline: false,
		}),
		Placeholder.configure({placeholder: () => props.placeholder}),
		Extension.create({
			name: 'quickAddHighlighter',
			addProseMirrorPlugins: () => [
				createQuickAddHighlighter(() => ({
					mode: authStore.settings.frontendSettings.quickAddMagicMode,
					labels: labels.value,
					projectExists,
				})),
			],
		}),
		Extension.create({
			name: 'quickAddKeys',
			priority: 1000,
			addKeyboardShortcuts: () => ({
				'Enter': ({editor}) => {
					if (editor.view.composing) return false
					emit('submit')
					return true
				},
				'Shift-Enter': ({editor}) => editor.commands.splitBlock(),
				'Escape': ({editor}) => editor.commands.blur(),
			}),
		}),
	],
	onCreate: ({editor}) => {
		editor.commands.setContent(toDoc(modelValue.value), {emitUpdate: false})
		// v-focus would fire before TipTap has mounted its contenteditable.
		if (props.autofocus && isAutoFocusViewport()) {
			editor.commands.focus('end')
		}
	},
	onUpdate: ({editor}) => {
		modelValue.value = editor.getText({blockSeparator: '\n'})
	},
})

watch(modelValue, value => {
	if (!editor.value || editor.value.getText({blockSeparator: '\n'}) === value) return
	editor.value.commands.setContent(toDoc(value), {emitUpdate: false})
})

watch([labels, () => projectList.projects], () => {
	editor.value?.view.dispatch(editor.value.state.tr.setMeta(quickAddHighlighterKey, true))
})

onBeforeUnmount(() => editor.value?.destroy())

defineExpose({
	focus: () => editor.value?.commands.focus(),
})
</script>

<template>
	<EditorContent
		:editor="editor"
		class="quick-add-input"
	/>
</template>

<style lang="scss" scoped>
.quick-add-input {
	:deep(.ProseMirror) {
		min-block-size: 2.5em;
		padding-block: calc(.5em - 1px);
		padding-inline: 2.5rem;
		border: 1px solid var(--input-border-color);
		border-radius: var(--input-radius);
		background: var(--white);
		line-height: 1.5;
		white-space: pre-wrap;
		outline: none;
		transition: border-color $transition;

		&:focus {
			border-color: var(--primary);
		}

		p.is-editor-empty:first-child::before {
			content: attr(data-placeholder);
			float: inline-start;
			block-size: 0;
			color: var(--grey-500);
			pointer-events: none;
		}
	}

	:deep(.quick-add-token) {
		border-radius: $radius;
		padding: .125rem .25rem;
		background: var(--grey-100);
	}

	:deep(.quick-add-token.is-new) {
		outline: 1px dashed var(--grey-400);
	}

	:deep(.quick-add-token.is-project) {
		color: var(--primary);
	}

	:deep(.quick-add-token.is-unknown) {
		color: var(--danger);
		text-decoration: line-through;
	}

	:deep(.quick-add-token.is-priority) {
		color: var(--danger);
		font-weight: 600;
	}

	:deep(.quick-add-token.is-date),
	:deep(.quick-add-token.is-repeat) {
		background: var(--primary);
		color: var(--white);
	}
}
</style>
