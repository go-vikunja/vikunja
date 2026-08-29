import {nextTick, toValue, type MaybeRefOrGetter, type Ref} from 'vue'

import StarterKit from '@tiptap/starter-kit'
import {Extension, mergeAttributes, type Editor, type Extensions} from '@tiptap/core'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {marked} from 'marked'

import Link from '@tiptap/extension-link'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import {Table, TableRow, TableCell, TableHeader} from '@tiptap/extension-table'
import Typography from '@tiptap/extension-typography'
import Image from '@tiptap/extension-image'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import Underline from '@tiptap/extension-underline'
import {Placeholder} from '@tiptap/extensions'
import HardBreak from '@tiptap/extension-hard-break'

import {TaskList} from '@tiptap/extension-list'
import {TaskItemWithId} from './taskItemWithId'
import {ListKeymapWithJoin} from './listKeymapWithJoin'
import {BlockquoteWithCommentId} from './blockquoteWithCommentId'
import {TaskLink, LINK_HTML_ATTRIBUTES} from './taskLink'

import Commands from './commands'
import suggestionSetup from './suggestion'
import {EmojiExtension} from './emoji/emojiExtension'

import {common, createLowlight} from 'lowlight'

import type {UploadCallback} from './types'
import type {ITask} from '@/modelTypes/ITask'
import type {IAttachment} from '@/modelTypes/IAttachment'
import AttachmentModel from '@/models/attachment'
import type AttachmentService from '@/services/attachment'

type CacheKey = `${ITask['id']}-${IAttachment['id']}`

export interface EditorExtensionDeps {
	t: (key: string) => string
	isEditing: Ref<boolean>
	isEditEnabled: () => boolean
	placeholder: MaybeRefOrGetter<string>
	contentHasChanged: Ref<boolean>
	bubbleSave: () => void
	getEditor: () => Editor | undefined
	uploadCallback: MaybeRefOrGetter<UploadCallback | undefined>
	uploadAndInsertFiles: (files: File[] | FileList) => void
	loadedAttachments: Ref<Record<string, string>>
	attachmentService: AttachmentService
}

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

// prevent links from extending after space
const NonInclusiveLink = Link.extend({
	inclusive() {
		return false
	},
})

const additionalLinkProtocols = [
	'ftp',
	'git',
	'obsidian',
	'notion',
	'message',
]

export function createEditorExtensions(deps: EditorExtensionDeps): Extensions {
	const {
		t,
		isEditing,
		isEditEnabled,
		placeholder,
		contentHasChanged,
		bubbleSave,
		getEditor,
		uploadCallback,
		uploadAndInsertFiles,
		loadedAttachments,
		attachmentService,
	} = deps

	// ProseMirror can call renderHTML twice per node on mount
	const inFlightBlobFetches = new Map<CacheKey, Promise<string>>()

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
					// never trust stored ids: a planted one would hijack the blob lookup
					parseHTML: () => null,
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

					// no live view: fail closed, never fall back to document
					const root = getEditor()?.view?.dom
					if (!root) return

					const img = root.querySelector(`[id="${id}"]`)

					if (!img || !(img instanceof HTMLImageElement)) return

					if (typeof loadedAttachments.value[cacheKey] === 'undefined') {
						let fetchPromise = inFlightBlobFetches.get(cacheKey)

						if (!fetchPromise) {
							const attachment = new AttachmentModel({taskId: taskId, id: attachmentId})
							fetchPromise = attachmentService.getBlobUrl(attachment) as Promise<string>
							inFlightBlobFetches.set(cacheKey, fetchPromise)
						}

						try {
							loadedAttachments.value[cacheKey] = await fetchPromise
						} catch {
							return
						} finally {
							// clear on failure too, else the rejected promise rethrows forever
							if (inFlightBlobFetches.get(cacheKey) === fetchPromise) {
								inFlightBlobFetches.delete(cacheKey)
							}
						}
					}

					img.src = loadedAttachments.value[cacheKey] as string
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

	const PasteHandler = Extension.create({
		name: 'pasteHandler',

		addProseMirrorPlugins() {
			return [
				new Plugin({
					key: new PluginKey('pasteHandler'),
					props: {
						handlePaste: (view, event) => {

							// Handle images pasted from clipboard
							if (typeof toValue(uploadCallback) !== 'undefined' && event.clipboardData?.items?.length) {

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

							// Don't convert markdown when pasting inside a code block
							const $from = view.state.selection.$from
							if ($from.parent.type.name === 'codeBlock') {
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

	return [
		// Starterkit:
		StarterKit.configure({
			codeBlock: false,
			hardBreak: false,
			blockquote: false,
			listKeymap: false,
			// Registered separately below with custom options; the StarterKit copies
			// would otherwise also run (e.g. link's openOnClick opening tabs in edit mode).
			link: false,
			underline: false,
		}),
		ListKeymapWithJoin,
		BlockquoteWithCommentId,

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
			placeholder({editor}) {
				if (!isEditing.value || editor.getText() !== '' && !editor.isFocused) {
					return ''
				}

				return toValue(placeholder) || t('input.editor.placeholder')
			},
		}),
		Typography,
		Subscript,
		Superscript,
		Underline,
		NonInclusiveLink.configure({
			openOnClick: false,
			validate: (href) => (new RegExp(
				`^(https?|${additionalLinkProtocols.join('|')}):\\/\\/`,
				'i',
			)).test(href),
			protocols: additionalLinkProtocols,
			HTMLAttributes: LINK_HTML_ATTRIBUTES,
		}),
		TaskLink,
		Table.configure({
			resizable: true,
		}),
		TableRow,
		TableHeader,
		// Custom TableCell with backgroundColor attribute
		CustomTableCell,

		CustomImage,

		TaskList,
		TaskItemWithId.configure({
			nested: true,
			onReadOnlyChecked(node, checked) {
				if (!isEditEnabled()) {
					return false
				}

				// Use taskId attribute to reliably find the correct node
				// This fixes GitHub issues #293 and #563
				const targetTaskId = node.attrs.taskId

				if (!targetTaskId) {
					// Fallback to original behavior if no ID (shouldn't happen)
					console.warn('TaskItem missing taskId, falling back to node comparison')
					getEditor()!.state.doc.descendants((subnode, pos) => {
						if (subnode === node) {
							const {tr} = getEditor()!.state
							tr.setNodeMarkup(pos, undefined, {
								...node.attrs,
								checked,
							})
							getEditor()!.view.dispatch(tr)
							bubbleSave()
						}
					})
					return true
				}

				// Find node by taskId for reliable matching
				getEditor()!.state.doc.descendants((subnode, pos) => {
					if (subnode.type.name === 'taskItem' && subnode.attrs.taskId === targetTaskId) {
						const {tr} = getEditor()!.state
						tr.setNodeMarkup(pos, undefined, {
							...subnode.attrs,
							checked,
						})
						getEditor()!.view.dispatch(tr)
						bubbleSave()
						return false // Stop iteration once found
					}
				})

				return true
			},
		}),

		Commands.configure({
			suggestion: suggestionSetup(t),
		}),

		EmojiExtension,

		PasteHandler,
	]
}
