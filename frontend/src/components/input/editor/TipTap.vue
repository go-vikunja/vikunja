<template>
	<div
		ref="tiptapInstanceRef"
		class="tiptap"
	>
		<EditorToolbar
			v-if="editor && isEditing"
			:editor="editor"
			@imageUploadClicked="triggerImageInput"
		/>
		<BubbleMenu
			v-if="editor && isEditing"
			:editor="editor"
		>
			<div class="editor-bubble__wrapper">
				<BaseButton
					v-tooltip="$t('input.editor.bold')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('bold') }"
					@click="() => editor?.chain().focus().toggleBold().run()"
				>
					<Icon :icon="['fa', 'fa-bold']" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('input.editor.italic')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('italic') }"
					@click="() => editor?.chain().focus().toggleItalic().run()"
				>
					<Icon :icon="['fa', 'fa-italic']" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('input.editor.underline')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('underline') }"
					@click="() => editor?.chain().focus().toggleUnderline().run()"
				>
					<Icon :icon="['fa', 'fa-underline']" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('input.editor.strikethrough')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('strike') }"
					@click="() => editor?.chain().focus().toggleStrike().run()"
				>
					<Icon :icon="['fa', 'fa-strikethrough']" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('input.editor.code')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('code') }"
					@click="() => editor?.chain().focus().toggleCode().run()"
				>
					<Icon :icon="['fa', 'fa-code']" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('input.editor.link')"
					class="editor-bubble__button"
					:class="{ 'is-active': editor.isActive('link') }"
					@click="setLink"
				>
					<Icon :icon="['fa', 'fa-link']" />
				</BaseButton>
			</div>
		</BubbleMenu>

		<EditorContent
			class="tiptap__editor"
			:class="{'tiptap__editor-is-edit-enabled': isEditing}"
			:editor="editor"
			@dblclick="setEditIfApplicable()"
			@click="focusIfEditing()"
		/>

		<input
			v-if="isEditing"
			id="tiptap__image-upload"
			ref="uploadInputRef"
			type="file"
			class="is-hidden"
			@change="addImage"
		>

		<ul
			v-if="bottomActions.length === 0 && !isEditing && isEditEnabled"
			class="tiptap__editor-actions d-print-none"
		>
			<li>
				<BaseButton
					class="done-edit"
					@click="() => setEdit()"
				>
					{{ $t('input.editor.edit') }}
				</BaseButton>
			</li>
		</ul>
		<ul
			v-if="bottomActions.length > 0"
			class="tiptap__editor-actions d-print-none"
		>
			<li v-if="isEditing && showSave">
				<BaseButton
					class="done-edit"
					@click="bubbleSave"
				>
					{{ $t('misc.save') }}
				</BaseButton>
			</li>
			<li v-if="!isEditing">
				<BaseButton
					class="done-edit"
					@click="() => setEdit()"
				>
					{{ $t('input.editor.edit') }}
				</BaseButton>
			</li>
			<li
				v-for="(action, k) in bottomActions"
				:key="k"
			>
				<BaseButton @click="action.action">
					{{ action.title }}
				</BaseButton>
			</li>
		</ul>
		<XButton
			v-else-if="isEditing && showSave"
			v-cy="'saveEditor'"
			class="mbs-4"
			variant="secondary"
			:shadow="false"
			:disabled="!contentHasChanged"
			@click="bubbleSave"
		>
			{{ $t('misc.save') }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {eventToHotkeyString} from '@github/hotkey'

import EditorToolbar from './EditorToolbar.vue'

import StarterKit from '@tiptap/starter-kit'
import {Extension, mergeAttributes} from '@tiptap/core'
import {EditorContent, type Extensions, useEditor} from '@tiptap/vue-3'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {marked} from 'marked'
import {BubbleMenu} from '@tiptap/vue-3/menus'

import Link from '@tiptap/extension-link'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import {Table, TableRow, TableCell, TableHeader} from '@tiptap/extension-table'
import Typography from '@tiptap/extension-typography'
import Image from '@tiptap/extension-image'
import Underline from '@tiptap/extension-underline'
import {Placeholder} from '@tiptap/extensions'

import {TaskItem, TaskList} from '@tiptap/extension-list'
import HardBreak from '@tiptap/extension-hard-break'

import {Node} from '@tiptap/pm/model'

import Commands from './commands'
import suggestionSetup from './suggestion'

import {common, createLowlight} from 'lowlight'

import type {BottomAction, UploadCallback} from './types'
import type {ITask} from '@/modelTypes/ITask'
import type {IAttachment} from '@/modelTypes/IAttachment'
import AttachmentModel from '@/models/attachment'
import AttachmentService from '@/services/attachment'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'

import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'
import inputPrompt from '@/helpers/inputPrompt'
import {setLinkInEditor} from '@/components/input/editor/setLinkInEditor'

const props = withDefaults(defineProps<{
	modelValue: string,
	uploadCallback?: UploadCallback,
	isEditEnabled?: boolean,
	bottomActions?: BottomAction[],
	showSave?: boolean,
	placeholder?: string,
	editShortcut?: string,
	enableDiscardShortcut?: boolean,
}>(), {
	uploadCallback: undefined,
	isEditEnabled: true,
	bottomActions: () => [],
	showSave: false,
	placeholder: '',
	editShortcut: '',
	enableDiscardShortcut: false,
})

const emit = defineEmits(['update:modelValue', 'save'])

const tiptapInstanceRef = ref<HTMLInputElement | null>(null)

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

type CacheKey = `${ITask['id']}-${IAttachment['id']}`
const loadedAttachments = ref<{
	[key: CacheKey]: string
}>({})

const CustomImage = Image.extend({
	addAttributes() {
		return {
			src: {
				default: null,
			},
			alt: {
				default: null,
			},
			title: {
				default: null,
			},
			id: {
				default: null,
			},
			'data-src': {
				default: null,
			},
		}
	},
	renderHTML({HTMLAttributes}) {
		if (HTMLAttributes.src?.startsWith(window.API_URL) || HTMLAttributes['data-src']?.startsWith(window.API_URL)) {
			const imageUrl = HTMLAttributes['data-src'] ?? HTMLAttributes.src

			// The url is something like /tasks/<id>/attachments/<id>
			const parts = imageUrl.slice(window.API_URL.length + 1).split('/')
			const taskId = Number(parts[1])
			const attachmentId = Number(parts[3])
			const cacheKey: CacheKey = `${taskId}-${attachmentId}`
			const id = 'tiptap-image-' + cacheKey

			nextTick(async () => {

				const img = document.getElementById(id)

				if (!img) return

				if (typeof loadedAttachments.value[cacheKey] === 'undefined') {

					const attachment = new AttachmentModel({taskId: taskId, id: attachmentId})

					const attachmentService = new AttachmentService()
					loadedAttachments.value[cacheKey] = await attachmentService.getBlobUrl(attachment)
				}

				img.src = loadedAttachments.value[cacheKey]
			})

			return ['img', mergeAttributes(this.options.HTMLAttributes, {
				'data-src': imageUrl,
				src: '#',
				alt: HTMLAttributes.alt,
				title: HTMLAttributes.title,
				id,
			})]
		}

		return ['img', mergeAttributes(this.options.HTMLAttributes, HTMLAttributes)]
	},
})

// prevent links from extending after space
const NonInclusiveLink = Link.extend({
	inclusive() {
		return false
	},
})

type Mode = 'edit' | 'preview'

const internalMode = ref<Mode>('preview')
const isEditing = computed(() => internalMode.value === 'edit' && props.isEditEnabled)
const contentHasChanged = ref<boolean>(false)

// TipTap crashes when inserting an image into an empty editor.
// To work around this, we're inserting an element first, then insert the image, then remove the element.
const UPLOAD_PLACEHOLDER_ELEMENT = '<p>UPLOAD_PLACEHOLDER</p>'

let lastSavedState = ''

watch(
	() => props.modelValue,
	(newValue) => {
		if (!contentHasChanged.value) {
			lastSavedState = newValue
		}
	},
	{ immediate: true },
)

watch(
	() => internalMode.value,
	mode => {
		if (mode === 'preview') {
			contentHasChanged.value = false
		}
	},
)

const additionalLinkProtocols = [
	'ftp',
	'git',
	'obsidian',
	'notion',
	'message',
]

const PasteHandler = Extension.create({
	name: 'pasteHandler',

	addProseMirrorPlugins() {
		return [
			new Plugin({
				key: new PluginKey('pasteHandler'),
				props: {
					handlePaste: (view, event) => {
						
						// Handle images pasted from clipboard
						if (typeof props.uploadCallback !== 'undefined' && event.clipboardData?.items?.length > 0) {

							for (const item of event.clipboardData.items) {
								if (item.kind === 'file' && item.type.startsWith('image/')) {
									const file = item.getAsFile()
									if (file) {
										uploadAndInsertFiles([file])
										return true
									}
								}
							}
						}
						
						const text = event.clipboardData?.getData('text/plain') || ''
						if (!text) {
							return false
						}

						const hasMarkdownSyntax = new RegExp('[*`_\\[\\]#-]').test(text)
						if (!hasMarkdownSyntax) {
							return false
						}

						const html = marked.parse(text)

						this.editor.commands.insertContent(html)
						return true
					},
				},
			}),
		]
	},
})


const extensions : Extensions = [
	// Starterkit:
	StarterKit.configure({
		codeBlock: false,
		hardBreak: false,
	}),

	CodeBlockLowlight.configure({
		lowlight: createLowlight(common),
	}),
	HardBreak.extend({
		addKeyboardShortcuts() {
			return {
				'Shift-Enter': () => this.editor.commands.setHardBreak(),
				'Mod-Enter': () => {
					if (contentHasChanged.value) {
						bubbleSave()
					}
					return true
				},
			}
		},
	}),

	Placeholder.configure({
		placeholder: ({editor}) => {
			if (!isEditing.value) {
				return ''
			}

			if (editor.getText() !== '' && !editor.isFocused) {
				return ''
			}

			return props.placeholder !== ''
				? props.placeholder
				: t('input.editor.placeholder')
		},
	}),
	Typography,
	Underline,
	NonInclusiveLink.configure({
		openOnClick: false,
		validate: (href: string) => (new RegExp(
			`^(https?|${additionalLinkProtocols.join('|')}):\\/\\/`,
			'i',
		)).test(href),
		protocols: additionalLinkProtocols,
	}),
	Table.configure({
		resizable: true,
	}),
	TableRow,
	TableHeader,
	// Custom TableCell with backgroundColor attribute
	CustomTableCell,

	CustomImage,

	TaskList,
	TaskItem.configure({
		nested: true,
		onReadOnlyChecked: (node: Node, checked: boolean): boolean => {
			if (!props.isEditEnabled) {
				return false
			}

			// The following is a workaround for this bug:
			// https://github.com/ueberdosis/tiptap/issues/4521
			// https://github.com/ueberdosis/tiptap/issues/3676

			editor.value!.state.doc.descendants((subnode, pos) => {
				if (subnode === node) {
					const {tr} = editor.value!.state
					tr.setNodeMarkup(pos, undefined, {
						...node.attrs,
						checked,
					})
					editor.value!.view.dispatch(tr)
					bubbleSave()
				}
			})


			return true
		},
	}),

	Commands.configure({
		suggestion: suggestionSetup(t),
	}),

	PasteHandler,
]

// Add a custom extension for the Escape key
if (props.enableDiscardShortcut) {
	extensions.push(Extension.create({
		name: 'escapeKey',

		addKeyboardShortcuts() {
			return {
				'Escape': () => {
					exitEditMode()
					return true
				},
			}
		},
	}))
}

const editor = useEditor({
	// eslint-disable-next-line vue/no-ref-object-reactivity-loss
	editable: isEditing.value,
	extensions: extensions,
	onUpdate: () => {
		bubbleNow()
	},
})

watch(
	() => isEditing.value,
	() => {
		editor.value?.setEditable(isEditing.value)
	},
	{immediate: true},
)

watch(
	() => props.modelValue,
	value => {
		if (!editor?.value) return

		if (editor.value.getHTML() === value) {
			return
		}

		setModeAndValue(value)
	},
	{immediate: true},
)

function bubbleNow() {
	if (editor.value?.getHTML() === props.modelValue ||
		(editor.value?.getHTML() === '<p></p>') && props.modelValue === '') {
		return
	}

	contentHasChanged.value = true
	emit('update:modelValue', editor.value?.getHTML())
}

function bubbleSave() {
	bubbleNow()
	lastSavedState = editor.value?.getHTML() ?? ''
	emit('save', lastSavedState)
	if (isEditing.value) {
		internalMode.value = 'preview'
	}
}

function exitEditMode() {
	editor.value?.commands.setContent(lastSavedState, false)
	if (isEditing.value) {
		internalMode.value = 'preview'
	}
}

function setEditIfApplicable() {
	if (!props.isEditEnabled) return
	if (isEditing.value) return

	setEdit()
}

function setEdit(focus: boolean = true) {
	internalMode.value = 'edit'
	if (focus) {
		editor.value?.commands.focus()
	}
}

onBeforeUnmount(() => editor.value?.destroy())

const uploadInputRef = ref<HTMLInputElement | null>(null)

function uploadAndInsertFiles(files: File[] | FileList) {
	if (typeof props.uploadCallback === 'undefined') {
		throw new Error('Can\'t add files here')
	}

	props.uploadCallback(files).then(urls => {
		urls?.forEach(url => {
			if (editor.value?.isEmpty) {
				editor.value
					?.chain()
					.focus()
					.insertContent(UPLOAD_PLACEHOLDER_ELEMENT)
					.run()
			}
			editor.value
				?.chain()
				.focus()
				.setImage({src: url})
				.run()
		})
		
		const html = editor.value?.getHTML().replace(UPLOAD_PLACEHOLDER_ELEMENT, '') ?? ''
		
		editor.value?.commands.setContent(html, false)
		
		bubbleSave()
	})
}

function triggerImageInput(event) {
	if (typeof props.uploadCallback !== 'undefined') {
		uploadInputRef.value?.click()
		return
	}

	addImage(event)
}

async function addImage(event) {

	if (typeof props.uploadCallback !== 'undefined') {
		const files = uploadInputRef.value?.files

		if (!files || files.length === 0) {
			return
		}

		uploadAndInsertFiles(files)

		return
	}

	const url = await inputPrompt(event.target.getBoundingClientRect())

	if (url) {
		editor.value?.chain().focus().setImage({src: url}).run()
		bubbleSave()
	}
}

function setLink(event) {
	setLinkInEditor(event.target.getBoundingClientRect(), editor.value)
}

onMounted(async () => {
	if (props.editShortcut !== '') {
		document.addEventListener('keydown', setFocusToEditor)
	}

	await nextTick()

	setModeAndValue(props.modelValue)
})

onBeforeUnmount(() => {
	if (props.editShortcut !== '') {
		document.removeEventListener('keydown', setFocusToEditor)
	}
})

function setModeAndValue(value: string) {
	internalMode.value = isEditorContentEmpty(value) ? 'edit' : 'preview'
	editor.value?.commands.setContent(value, false)
}


// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
function setFocusToEditor(event) {
	if (event.target.shadowRoot) {
		return
	}

	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return
	if (hotkeyString !== props.editShortcut ||
		event.target.tagName.toLowerCase() === 'input' ||
		event.target.tagName.toLowerCase() === 'textarea' ||
		event.target.contentEditable === 'true') {
		return
	}

	event.preventDefault()

	if (!isEditing.value && props.isEditEnabled) {
		internalMode.value = 'edit'
	}

	editor.value?.commands.focus()
}

function focusIfEditing() {
	if (isEditing.value) {
		editor.value?.commands.focus()
	}
}

function clickTasklistCheckbox(event) {
	event.stopImmediatePropagation()

	if (event.target.localName !== 'p') {
		return
	}

	event.target.parentNode.parentNode.firstChild.click()
}

watch(
	() => isEditing.value,
	async editing => {
		await nextTick()

		let checkboxes = tiptapInstanceRef.value?.querySelectorAll('[data-checked]')
		if (typeof checkboxes === 'undefined' || checkboxes.length === 0) {
			// For some reason, this works when we check a second time.
			await nextTick()

			checkboxes = tiptapInstanceRef.value?.querySelectorAll('[data-checked]')
			if (typeof checkboxes === 'undefined' || checkboxes.length === 0) {
				return
			}
		}

		if (editing) {
			checkboxes.forEach(check => {
				if (check.children.length < 2) {
					return
				}

				// We assume the first child contains the label element with the checkbox and the second child the actual label
				// When the actual label is clicked, we forward that click to the checkbox.
				check.children[1].removeEventListener('click', clickTasklistCheckbox)
			})

			return
		}

		checkboxes.forEach(check => {
			if (check.children.length < 2) {
				return
			}

			// We assume the first child contains the label element with the checkbox and the second child the actual label
			// When the actual label is clicked, we forward that click to the checkbox.
			check.children[1].removeEventListener('click', clickTasklistCheckbox)
			check.children[1].addEventListener('click', clickTasklistCheckbox)
		})
	},
	{immediate: true},
)
</script>

<style lang="scss">
.tiptap__editor {
	transition: box-shadow $transition;
	border-radius: $radius;
	
	&.tiptap__editor-is-edit-enabled {
		min-block-size: 10rem;

		.ProseMirror {
			padding: .5rem;
		}

		&:focus-within, &:focus {
			box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);
		}

		ul[data-type='taskList'] li > div {
			cursor: text;
		}
	}
}

.tiptap p::before {
	content: attr(data-placeholder);
	color: var(--grey-400);
	pointer-events: none;
	block-size: 0;
	float: inline-start;
}

// Basic editor styles
.ProseMirror {
	padding: .5rem .5rem .5rem 0;

	&:focus-within, &:focus {
		box-shadow: none;
	}

	> * + * {
		margin-block-start: 0.75em;
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

	code {
		background-color: var(--grey-200);
		color: var(--grey-700);
		border-radius: $radius;
	}

	pre {
		background: var(--grey-200);
		color: var(--grey-700);
		font-family: JetBrainsMono, monospace;
		padding: 0.75rem 1rem;
		border-radius: $radius;

		code {
			color: inherit;
			padding: 0;
			background: none;
			font-size: 0.8rem;
		}

		.hljs-comment,
		.hljs-quote {
			color: var(--grey-500);
		}

		.hljs-variable,
		.hljs-template-variable,
		.hljs-attribute,
		.hljs-tag,
		.hljs-name,
		.hljs-regexp,
		.hljs-link,
		.hljs-selector-id,
		.hljs-selector-class {
			color: var(--code-variable);
		}

		.hljs-number,
		.hljs-meta,
		.hljs-built_in,
		.hljs-builtin-name,
		.hljs-literal,
		.hljs-type,
		.hljs-params {
			color: var(--code-literal);
		}

		.hljs-string,
		.hljs-symbol,
		.hljs-bullet {
			color: var(--code-symbol);
		}

		.hljs-title,
		.hljs-section {
			color: var(--code-section);
		}

		.hljs-keyword,
		.hljs-selector-tag {
			color: var(--code-keyword);
		}

		.hljs-emphasis {
			font-style: italic;
		}

		.hljs-strong {
			font-weight: 700;
		}
	}

	img {
		max-inline-size: 100%;
		block-size: auto;

		&.ProseMirror-selectednode {
			outline: 3px solid var(--primary);
		}
	}

	blockquote {
		padding-inline-start: 1rem;
		border-inline-start: 2px solid rgba(#0d0d0d, 0.1);
	}

	hr {
		border: none;
		border-block-start: 2px solid rgba(#0d0d0d, 0.1);
		margin: 2rem 0;
	}
	
	// Table-specific styling
	table {
		border-collapse: collapse;
		table-layout: fixed;
		inline-size: 100%;
		margin: 0;
		overflow: hidden;

		td,
		th {
			min-inline-size: 1em;
			border: 2px solid var(--grey-300) !important;
			padding: 3px 5px;
			vertical-align: top;
			box-sizing: border-box;
			position: relative;

			> * {
				margin-block-end: 0;
			}
		}

		th {
			font-weight: bold;
			text-align: start;
			background-color: var(--grey-200);
		}

		.selectedCell:after {
			z-index: 2;
			position: absolute;
			content: '';
			inset-inline-start: 0;
			inset-inline-end: 0;
			inset-block-start: 0;
			inset-block-end: 0;
			background: rgba(200, 200, 255, 0.4);
			pointer-events: none;
		}

		.column-resize-handle {
			position: absolute;
			inset-inline-end: -2px;
			inset-block-start: 0;
			inset-block-end: -2px;
			inline-size: 4px;
			background-color: #aaddff;
			pointer-events: none;
		}

		p {
			margin: 0;
		}
	}

	// Lists

	ul {
		margin-inline-start: .5rem;
		margin-block-start: 0 !important;

		li {
			margin-block-start: 0;
		}

		p {
			margin-block-end: 0 !important;
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
	margin-inline-start: 0;

	li[data-checked='true'] {
		color: var(--grey-500);
		text-decoration: line-through;
	}

	li {
		display: flex;
		margin-block-start: 0.25rem;

		> label {
			flex: 0 0 auto;
			margin-inline-end: 0.5rem;
			user-select: none;
		}

		> div {
			flex: 1 1 auto;
			cursor: pointer;
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
		inline-size: 2rem;
		block-size: 2rem;
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
