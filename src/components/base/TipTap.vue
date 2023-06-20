<template>
	<div class="tiptap">
		<EditorToolbar v-if="editor" :editor="editor" />
		<editor-content class="tiptap__editor" :editor="editor" />
	</div>
</template>

<script lang="ts">
export const TIPTAP_TEXT_VALUE_PREFIX = '<!-- VIKUNJA TIPTAP -->\n'
const tiptapRegex = new RegExp(`${TIPTAP_TEXT_VALUE_PREFIX}`, 's')
</script>

<script setup lang="ts">
import {ref, watch, computed, onBeforeUnmount, type PropType} from 'vue'
import {marked} from 'marked'
import {refDebounced} from '@vueuse/core'

import EditorToolbar from './EditorToolbar.vue'

import Link from '@tiptap/extension-link'

import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import Table from '@tiptap/extension-table'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import TableRow from '@tiptap/extension-table-row'
import Highlight from '@tiptap/extension-highlight'
import Typography from '@tiptap/extension-typography'
import Document from '@tiptap/extension-document'
import Image from '@tiptap/extension-image'
// import Text from '@tiptap/extension-text'

import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'

import CharacterCount from '@tiptap/extension-character-count'

import StarterKit from '@tiptap/starter-kit'
import {EditorContent, useEditor, VueNodeViewRenderer} from '@tiptap/vue-3'

// load all highlight.js languages
import {lowlight} from 'lowlight'

import CodeBlock from './CodeBlock.vue'

// const CustomDocument = Document.extend({
// 	content: 'taskList',
// })

const CustomTaskItem = TaskItem.configure({
	nested: true,
})

const CustomTableCell = TableCell.extend({
	addAttributes() {
		return {
			// extend the existing attributes …
			...this.parent?.(),

			// and add a new one …
			backgroundColor: {
				default: null,
				parseHTML: (element: HTMLElement) => element.getAttribute('data-background-color'),
				renderHTML: (attributes) => {
					return {
						'data-background-color': attributes.backgroundColor,
						style: `background-color: ${attributes.backgroundColor}`,
					}
				},
			},
		}
	},
})

const props = withDefaults(defineProps<{
	modelValue?: string,
}>(), {
	modelValue: '',
})

const emit = defineEmits(['update:modelValue', 'change'])

const inputHTML = ref('')
watch(
	() => props.modelValue,
	() => {
		if (!props.modelValue.startsWith(TIPTAP_TEXT_VALUE_PREFIX)) {
			// convert Markdown to HTML
			return TIPTAP_TEXT_VALUE_PREFIX + marked.parse(props.modelValue)
		}

		return props.modelValue.replace(tiptapRegex, '')
	},
	{ immediate: true },
)

const debouncedInputHTML = refDebounced(inputHTML, 1000)

watch(debouncedInputHTML, (value) => {
	emit('update:modelValue', TIPTAP_TEXT_VALUE_PREFIX + value)
	emit('change', TIPTAP_TEXT_VALUE_PREFIX + value) // FIXME: remove this
})

const editor = useEditor({
	content: inputHTML.value,
	extensions: [
		StarterKit,
		Highlight,
		Typography,
		Link.configure({
			openOnClick: false,
			validate: (href: string) => /^https?:\/\//.test(href),
		}),
		// Table.configure({
		// 	resizable: true,
		// }),
		// TableRow,
		// TableHeader,
		// // Default TableCell
		// // TableCell,
		// // Custom TableCell with backgroundColor attribute
		// CustomTableCell,

		// // start
		// Document,
		// // Text,
		// Image,

		// // Tasks
		// CustomDocument,
		TaskList,
		CustomTaskItem,

		// // character count
		// CharacterCount,

		// CodeBlockLowlight.extend({
		// 	addNodeView() {
		// 		return VueNodeViewRenderer(CodeBlock)
		// 	},
		// }).configure({ lowlight }),
	],
	onUpdate: () => {
		// HTML
		inputHTML.value = editor.value!.getHTML()

		// JSON
		// this.$emit('update:modelValue', this.editor.getJSON())
	},
})

watch(inputHTML, (value) => {
	if (!editor.value) return
	// HTML
	const isSame = editor.value.getHTML() === value

	// JSON
	// const isSame = JSON.stringify(editor.value.getJSON()) === JSON.stringify(value)

	if (isSame) {
		return
	}

	editor.value.commands.setContent(value, false)
})

onBeforeUnmount(() => editor.value?.destroy())
</script>

<style lang="scss">
.tiptap__editor {
	// box-sizing: border-box;
	// height: auto;
	min-height: 150px;
	border: 1px solid #ddd;
	border-bottom-left-radius: 4px;
	border-bottom-right-radius: 4px;
	padding: 10px;
	// font: inherit;
	// z-index: 0;
	// word-wrap: break-word;

	border: 1px solid var(--grey-200) !important;
	background: var(--white);
}

/* Basic editor styles */
.ProseMirror {
	> * + * {
		margin-top: 0.75em;
	}

	ul,
	ol {
		padding: 0 1rem;
	}

	h1,
	h2,
	h3,
	h4,
	h5,
	h6 {
		line-height: 1.1;
	}

	a {
		color: #68cef8;
	}

	code {
		background-color: rgba(#616161, 0.1);
		color: #616161;
	}

	pre {
		background: #0d0d0d;
		color: #fff;
		font-family: "JetBrainsMono", monospace;
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;

		code {
			color: inherit;
			padding: 0;
			background: none;
			font-size: 0.8rem;
		}
	}

	pre {
		background: #0d0d0d;
		color: #fff;
		font-family: "JetBrainsMono", monospace;
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;

		code {
			color: inherit;
			padding: 0;
			background: none;
			font-size: 0.8rem;
		}

		.hljs-comment,
		.hljs-quote {
			color: #616161;
		}

		.hljs-variable,
		.hljs-template-variable,
		.hljs-attribute,
		.hljs-tag,
		.hljs-name,
		.hljs-regexp,
		.hljs-link,
		.hljs-name,
		.hljs-selector-id,
		.hljs-selector-class {
			color: #f98181;
		}

		.hljs-number,
		.hljs-meta,
		.hljs-built_in,
		.hljs-builtin-name,
		.hljs-literal,
		.hljs-type,
		.hljs-params {
			color: #fbbc88;
		}

		.hljs-string,
		.hljs-symbol,
		.hljs-bullet {
			color: #b9f18d;
		}

		.hljs-title,
		.hljs-section {
			color: #faf594;
		}

		.hljs-keyword,
		.hljs-selector-tag {
			color: #70cff8;
		}

		.hljs-emphasis {
			font-style: italic;
		}

		.hljs-strong {
			font-weight: 700;
		}
	}

	img {
		max-width: 100%;
		height: auto;

		&.ProseMirror-selectednode {
			outline: 3px solid #68cef8;
		}
	}

	blockquote {
		padding-left: 1rem;
		border-left: 2px solid rgba(#0d0d0d, 0.1);
	}

	hr {
		border: none;
		border-top: 2px solid rgba(#0d0d0d, 0.1);
		margin: 2rem 0;
	}
}

/* Table-specific styling */
.ProseMirror {
	table {
		border-collapse: collapse;
		table-layout: fixed;
		width: 100%;
		margin: 0;
		overflow: hidden;

		td,
		th {
			min-width: 1em;
			border: 2px solid #ced4da;
			padding: 3px 5px;
			vertical-align: top;
			box-sizing: border-box;
			position: relative;

			> * {
				margin-bottom: 0;
			}
		}

		th {
			font-weight: bold;
			text-align: left;
			background-color: #f1f3f5;
		}

		.selectedCell:after {
			z-index: 2;
			position: absolute;
			content: "";
			left: 0;
			right: 0;
			top: 0;
			bottom: 0;
			background: rgba(200, 200, 255, 0.4);
			pointer-events: none;
		}

		.column-resize-handle {
			position: absolute;
			right: -2px;
			top: 0;
			bottom: -2px;
			width: 4px;
			background-color: #adf;
			pointer-events: none;
		}

		p {
			margin: 0;
		}
	}
}

.tableWrapper {
	overflow-x: auto;
}

.resize-cursor {
	cursor: ew-resize;
	cursor: col-resize;
}

// tasklist
ul[data-type="taskList"] {
	list-style: none;
	padding: 0;
	margin-left: 0;
	margin-top: 0;
	
	p {
		margin-bottom: 0 !important;
	}

	li {
		display: flex;

		> label {
			flex: 0 0 auto;
			margin-right: 0.5rem;
			user-select: none;
		}

		> div {
			flex: 1 1 auto;
		}
	}

	input[type="checkbox"] {
		cursor: pointer;
	}
}

// character count
.character-count {
	margin-top: 1rem;
	display: flex;
	align-items: center;
	color: #68cef8;

	&--warning {
		color: #fb5151;
	}

	&__graph {
		margin-right: 0.5rem;
	}

	&__text {
		color: #868e96;
	}
}
</style>
