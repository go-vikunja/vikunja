<template>
	<div class="tiptap">
		<EditorToolbar
			v-if="editor && isEditEnabled"
			:editor="editor"
			:upload-callback="uploadCallback"
		/>
		<BubbleMenu
			v-if="editor && isEditEnabled"
			:editor="editor"
			class="editor-bubble__wrapper"
		>
			<BaseButton
				class="editor-bubble__button"
				@click="editor.chain().focus().toggleBold().run()"
				:class="{ 'is-active': editor.isActive('bold') }"
				v-tooltip="$t('input.editor.bold')"
			>
				<icon :icon="['fa', 'fa-bold']"/>
			</BaseButton>
			<BaseButton
				class="editor-bubble__button"
				@click="editor.chain().focus().toggleItalic().run()"
				:class="{ 'is-active': editor.isActive('italic') }"
				v-tooltip="$t('input.editor.italic')"
			>
				<icon :icon="['fa', 'fa-italic']"/>
			</BaseButton>
			<BaseButton
				class="editor-bubble__button"
				@click="editor.chain().focus().toggleUnderline().run()"
				:class="{ 'is-active': editor.isActive('underline') }"
				v-tooltip="$t('input.editor.underline')"
			>
				<icon :icon="['fa', 'fa-underline']"/>
			</BaseButton>
			<BaseButton
				class="editor-bubble__button"
				@click="editor.chain().focus().toggleStrike().run()"
				:class="{ 'is-active': editor.isActive('strike') }"
				v-tooltip="$t('input.editor.strikethrough')"
			>
				<icon :icon="['fa', 'fa-strikethrough']"/>
			</BaseButton>
			<BaseButton
				class="editor-bubble__button"
				@click="editor.chain().focus().toggleCode().run()"
				:class="{ 'is-active': editor.isActive('code') }"
				v-tooltip="$t('input.editor.code')"
			>
				<icon :icon="['fa', 'fa-code']"/>
			</BaseButton>
			<BaseButton
				class="editor-bubble__button"
				@click="setLink"
				:class="{ 'is-active': editor.isActive('link') }"
				v-tooltip="$t('input.editor.link')"
			>
				<icon :icon="['fa', 'fa-link']"/>
			</BaseButton>
		</BubbleMenu>

		<editor-content
			class="tiptap__editor"
			:class="{'tiptap__editor-is-empty': isEmpty, 'tiptap__editor-is-edit-enabled': isEditEnabled}"
			:editor="editor"
		/>

		<input
			v-if="isEditEnabled"
			type="file"
			id="tiptap__image-upload"
			class="is-hidden"
			ref="uploadInputRef"
			@change="addImage"
		/>

		<ul class="tiptap__editor-actions d-print-none" v-if="bottomActions.length > 0">
			<li v-if="isEditEnabled && showSave">
				<BaseButton
					@click="bubbleSave"
					class="done-edit">
					{{ $t('misc.save') }}
				</BaseButton>
			</li>
			<li v-for="(action, k) in bottomActions" :key="k">
				<BaseButton @click="action.action">{{ action.title }}</BaseButton>
			</li>
		</ul>
		<x-button
			v-else-if="isEditEnabled && showSave"
			class="mt-4"
			@click="bubbleSave"
			variant="secondary"
			:shadow="false"
			v-cy="'saveEditor'"
		>
			{{ $t('misc.save') }}
		</x-button>
	</div>
</template>

<script lang="ts">
export const TIPTAP_TEXT_VALUE_PREFIX = '<!-- VIKUNJA TIPTAP -->\n'
const tiptapRegex = new RegExp(`${TIPTAP_TEXT_VALUE_PREFIX}`, 's')
</script>

<script setup lang="ts">
import {ref, watch, onBeforeUnmount, nextTick, onMounted, computed} from 'vue'
import {marked} from 'marked'
import {refDebounced} from '@vueuse/core'

import EditorToolbar from './EditorToolbar.vue'

import Link from '@tiptap/extension-link'

import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import Table from '@tiptap/extension-table'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import TableRow from '@tiptap/extension-table-row'
import Typography from '@tiptap/extension-typography'
import Image from '@tiptap/extension-image'
import Underline from '@tiptap/extension-underline'

import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'

import {Blockquote} from '@tiptap/extension-blockquote'
import {Bold} from '@tiptap/extension-bold'
import {BulletList} from '@tiptap/extension-bullet-list'
import {Code} from '@tiptap/extension-code'
import {CodeBlock} from '@tiptap/extension-code-block'
import {Document} from '@tiptap/extension-document'
import {Dropcursor} from '@tiptap/extension-dropcursor'
import {Gapcursor} from '@tiptap/extension-gapcursor'
import {HardBreak} from '@tiptap/extension-hard-break'
import {Heading} from '@tiptap/extension-heading'
import {History} from '@tiptap/extension-history'
import {HorizontalRule} from '@tiptap/extension-horizontal-rule'
import {Italic} from '@tiptap/extension-italic'
import {ListItem} from '@tiptap/extension-list-item'
import {OrderedList} from '@tiptap/extension-ordered-list'
import {Paragraph} from '@tiptap/extension-paragraph'
import {Strike} from '@tiptap/extension-strike'
import {Text} from '@tiptap/extension-text'
import {BubbleMenu, EditorContent, useEditor} from '@tiptap/vue-3'

import Commands from './commands'
import suggestionSetup from './suggestion'

// load all highlight.js languages
import {lowlight} from 'lowlight'

import type {BottomAction, UploadCallback} from './types'
import type {ITask} from '@/modelTypes/ITask'
import type {IAttachment} from '@/modelTypes/IAttachment'
import AttachmentModel from '@/models/attachment'
import AttachmentService from '@/services/attachment'
import {useI18n} from 'vue-i18n'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/button.vue'
import {Placeholder} from '@tiptap/extension-placeholder'
import {eventToHotkeyString} from '@github/hotkey'
import {useBaseStore} from '@/stores/base'

const {t} = useI18n()

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


const {
	modelValue,
	uploadCallback,
	isEditEnabled = true,
	bottomActions = [],
	showSave = false,
	placeholder = '',
	editShortcut = '',
} = defineProps<{
	modelValue: string,
	uploadCallback?: UploadCallback,
	isEditEnabled?: boolean,
	bottomActions?: BottomAction[],
	showSave?: boolean,
	placeholder?: string,
	editShortcut?: string,
}>()

const baseStore = useBaseStore()

const emit = defineEmits(['update:modelValue', 'save'])

const inputHTML = ref('')
watch(
	() => modelValue,
	() => {
		if (modelValue === '') {
			inputHTML.value = TIPTAP_TEXT_VALUE_PREFIX
			return
		}

		if (!modelValue.startsWith(TIPTAP_TEXT_VALUE_PREFIX)) {
			// convert Markdown to HTML
			inputHTML.value = TIPTAP_TEXT_VALUE_PREFIX + marked.parse(modelValue)
			nextTick(() => loadImages())
			return
		}

		inputHTML.value = modelValue.replace(tiptapRegex, '')
		nextTick(() => loadImages())
	},
	{immediate: true},
)

const isEmpty = computed(() => inputHTML.value === '')

function onImageAdded() {
	bubbleSave()
	loadImages()
}

type CacheKey = `${ITask['id']}-${IAttachment['id']}`
const loadedAttachments = ref<{ [key: CacheKey]: string }>({})

function loadImages() {
	const attachmentImage = document.querySelectorAll<HTMLImageElement>('.tiptap__editor img')
	const attachmentService = new AttachmentService()
	if (attachmentImage) {
		Array.from(attachmentImage).forEach(async (img) => {
			if (!img.src.startsWith(window.API_URL)) {
				return
			}
			// The url is something like /tasks/<id>/attachments/<id>
			const parts = img.src.slice(window.API_URL.length + 1).split('/')
			const taskId = Number(parts[1])
			const attachmentId = Number(parts[3])
			const cacheKey: CacheKey = `${taskId}-${attachmentId}`

			if (typeof loadedAttachments.value[cacheKey] !== 'undefined') {
				img.src = loadedAttachments.value[cacheKey]
				return
			}

			const attachment = new AttachmentModel({taskId: taskId, id: attachmentId})

			const url = await attachmentService.getBlobUrl(attachment)
			img.src = url
			loadedAttachments.value[cacheKey] = url
		})
	}

}

const debouncedInputHTML = refDebounced(inputHTML, 1000)

watch(debouncedInputHTML, () => bubbleNow())

function bubbleNow() {
	emit('update:modelValue', TIPTAP_TEXT_VALUE_PREFIX + inputHTML.value)
}

function bubbleSave() {
	bubbleNow()
	emit('save', TIPTAP_TEXT_VALUE_PREFIX + inputHTML.value)
}

const editor = useEditor({
	content: inputHTML.value,
	editable: isEditEnabled,
	extensions: [
		// Starterkit:
		Blockquote,
		Bold,
		BulletList,
		Code,
		CodeBlockLowlight.configure({
			lowlight,
		}),
		Document,
		Dropcursor,
		Gapcursor,
		HardBreak.extend({
			addKeyboardShortcuts() {
				return {
					'Mod-Enter': () => {
						bubbleSave()
					},
				}
			},
		}),
		Heading,
		History,
		HorizontalRule,
		Italic,
		ListItem,
		OrderedList,
		Paragraph,
		Strike,
		Text,

		Placeholder.configure({
			placeholder: ({editor}) => {
				if (!isEditEnabled) {
					return ''
				}

				if (editor.getText() !== '' && !editor.isFocused) {
					return ''
				}

				return placeholder !== ''
					? placeholder
					: t('input.editor.placeholder')
			},
		}),
		Typography,
		Underline,
		Link.configure({
			openOnClick: true,
			validate: (href: string) => /^https?:\/\//.test(href),
		}),
		Table.configure({
			resizable: true,
		}),
		TableRow,
		TableHeader,
		// Custom TableCell with backgroundColor attribute
		CustomTableCell,

		Image,

		TaskList,
		TaskItem.configure({
			nested: true,
		}),

		Commands
			.configure({suggestion: suggestionSetup(t)})
			.extend({name: 'slashMenuCommands'}),
		BubbleMenu,
	],
	onUpdate: () => {
		// HTML
		inputHTML.value = editor.value!.getHTML()

		// JSON
		// this.$emit('update:modelValue', this.editor.getJSON())
	},
	onFocus() {
		baseStore.setEditorFocused(true)
	},
	onBlur() {
		baseStore.setEditorFocused(false)
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

const uploadInputRef = ref<HTMLInputElement | null>(null)

function uploadAndInsertFiles(files: File[] | FileList) {
	uploadCallback(files).then(urls => {
		urls?.forEach(url => {
			editor.value
				?.chain()
				.focus()
				.setImage({src: url})
				.run()
		})
		onImageAdded()
	})
}

function addImage() {

	if (typeof uploadCallback !== 'undefined') {
		const files = uploadInputRef.value?.files

		if (!files || files.length === 0) {
			return
		}

		uploadAndInsertFiles(files)

		return
	}

	const url = window.prompt('URL')

	if (url) {
		editor.value?.chain().focus().setImage({src: url}).run()
		onImageAdded()
	}
}

function setLink() {
	const previousUrl = editor.value?.getAttributes('link').href
	const url = window.prompt('URL', previousUrl)

	// cancelled
	if (url === null) {
		return
	}

	// empty
	if (url === '') {
		editor.value
			?.chain()
			.focus()
			.extendMarkRange('link')
			.unsetLink()
			.run()

		return
	}

	// update link
	editor.value
		?.chain()
		.focus()
		.extendMarkRange('link')
		.setLink({href: url, target: '_blank'})
		.run()
}

onMounted(() => {
	document.addEventListener('paste', handleImagePaste)
	if (editShortcut !== '') {
		document.addEventListener('keydown', setFocusToEditor)
	}
})

onBeforeUnmount(() => {
	document.removeEventListener('paste', handleImagePaste)
	if (editShortcut !== '') {
		document.removeEventListener('keydown', setFocusToEditor)
	}
})

function handleImagePaste(event) {
	event.preventDefault()
	event?.clipboardData?.items?.forEach(i => {
		if (i.kind === 'file' && i.type.startsWith('image/')) {
			uploadAndInsertFiles([i.getAsFile()])
		}
	})
}

// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
function setFocusToEditor(event) {
	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return
	if (hotkeyString !== editShortcut || baseStore.editorFocused) return
	event.preventDefault()

	editor.value?.commands.focus()
}
</script>

<style lang="scss">
.tiptap__editor {
	min-height: 10rem;
	transition: box-shadow $transition;
	border-radius: $radius;

	&:focus-within, &:focus {
		box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);
	}
}

.tiptap p.is-empty::before {
	content: attr(data-placeholder);
	color: var(--grey-400);
	pointer-events: none;
	height: 0;
	float: left;
}

// Basic editor styles
.ProseMirror {
	padding: .5rem;

	&:focus-within, &:focus {
		box-shadow: none;
	}

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
		font-family: 'JetBrainsMono', monospace;
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
		font-family: 'JetBrainsMono', monospace;
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

.ProseMirror {
	/* Table-specific styling */
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
			content: '';
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

	// Lists
	ul {
		margin-left: .5rem;
		margin-top: 0 !important;

		li {
			margin-top: 0;
		}

		p {
			margin-bottom: 0 !important;
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
ul[data-type='taskList'] {
	list-style: none;
	padding: 0;
	margin-left: 0;

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

	input[type='checkbox'] {
		cursor: pointer;
	}
}

.editor-bubble__wrapper {
	background: var(--white);
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	box-shadow: var(--shadow-md);
	display: flex;
	overflow: hidden;
}

.editor-bubble__button {
	color: var(--grey-700);
	transition: all $transition;
	background: transparent;

	svg {
		box-sizing: border-box;
		display: block;
		width: 1rem;
		height: 1rem;
		padding: .5rem;
		margin: 0;
	}

	&:hover {
		background: var(--grey-200);
	}
}

ul.tiptap__editor-actions {
	font-size: .8rem;
	margin: 0;

	li {
		display: inline-block;

		&::after {
			content: '·';
			padding: 0 .25rem;
		}

		&:last-child:after {
			content: '';
		}
	}

	&, a {
		color: var(--grey-500);

		&.done-edit {
			color: var(--primary);
		}
	}

	a:hover {
		text-decoration: underline;
	}
}
</style>
